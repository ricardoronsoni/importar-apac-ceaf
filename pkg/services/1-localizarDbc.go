package services

import (
	"dados-apac/pkg/database"
	"dados-apac/pkg/utils"
	"fmt"
	"strconv"
	"time"

	"github.com/jlaffaye/ftp"
)

const (
	dbcTempPath    = "./dbc/dbcTemp/"
	dbfTempPath    = "./dbc/dbfTemp/"
	csvTempPath    = "./dbc/csvTemp/"
	dbcParaDbfPath = "./dbc/dbcParaDbf/"
)

var (
	data                 database.Database
	clientFtp            *ftp.ServerConn
	todosArquivos        []*ftp.Entry
	arquivosSelecionados []string
	inicioProcesso       time.Time
	qtdArquivoPersistido int
)

// IniciarMonitor inicia o processo de busca dos dados no Datasus
func IniciarMonitor() {
	inicioProcesso = time.Now()
	qtdArquivoPersistido = 0

	if err := criarConexoes(); err != nil {
		return
	}

	if err := localizarInconsistenciasBanco(); err != nil {
		return
	}

	if err := listarArquivosFtp(); err != nil {
		return
	}

	arquivosSelecionados = selecionarArquivosFtp(todosArquivos)
	emailArquivosLocalizados(arquivosSelecionados)
	iniciarProcessamento()

	if err := clientFtp.Quit(); err != nil {
		utils.Logger(err, "error")
	}
}

// localizarInconsistenciasBanco ira localizar no banco de dados arquivos que comecaram a ser importados e nao foram finalizados. Os dados desses arquivos serão exlcuidos do banco de dados.
func localizarInconsistenciasBanco() error {
	utils.Logger("Verificando a consistencia do banco de dados.", "info")

	arquivosErro, err := data.InconsistenciaAuditoria()
	if err != nil {
		utils.Logger(err, "fatal")
		return nil
	}

	if len(arquivosErro) == 0 {
		utils.Logger("Nao foram encontradas inconsistencias do banco de dados.", "info")
	} else {
		utils.Logger("Foi(ram) encontrado(s) "+fmt.Sprint(len(arquivosErro))+" arquivo(s) com inconsistencias do banco de dados.", "info")

		for _, arquivo := range arquivosErro {
			purgeInsertDbc(arquivo)
		}
	}

	return nil
}

// criarConexoes cria conexao com o banco de dados e com o FTP do Datasus
func criarConexoes() error {
	utils.Logger("Conectando com o servidor FTP do Datasus.", "info")
	db := database.DbConnection()
	data = database.Database{
		SqlDb: db,
	}

	var err error
	clientFtp, err = utils.FtpConnection()
	if err != nil {
		return err
	}

	return nil
}

// listarArquivosFtp ira listar todos os arquivos no diretorio FTP
func listarArquivosFtp() error {
	var err error
	todosArquivos, err = clientFtp.List(".")
	if err != nil {
		utils.Logger(err.Error(), "error")
		return err
	}

	return nil
}

// selecionarArquivosFtp seleciona: 1) se eh arquivo de medicamento; 2) se o arquivo segue o padrao AMufaamm.dbc; 3) se arquivo esta dentro do range de data para download
func selecionarArquivosFtp(todosArquivos []*ftp.Entry) []string {
	arquivosSelecionados := []string{}

	for _, entry := range todosArquivos {
		//verifica os arquivos de medicamento e se o arquivo segue o padrao AMufaamm.dbc
		if entry.Name[0:2] == "AM" && len(entry.Name) == 12 { //&& entry.Name[2:4] == "PA" {
			//verifica se o arquivo esta dentro do range de data para download
			var (
				competencia int
				err         error
			)
			if competencia, err = strconv.Atoi(entry.Name[4:8]); err != nil {
				utils.Logger(err, "error")
			}

			arquivoPersistido := true
			if competencia >= utils.DataInicial && competencia <= utils.DataFinal {
				arquivoPersistido = data.ConsultarArquivo(entry.Name)
			}

			if !arquivoPersistido {
				arquivosSelecionados = append(arquivosSelecionados, entry.Name)
				utils.Logger("Arquivo "+entry.Name+" localizado para importação.", "info")
			}
		}
	}

	return arquivosSelecionados
}

// iniciarProcessamento pega a relacao dos arquivos localizados para importacao e inicia o processo de importacao de cada um individualmente
func iniciarProcessamento() {
	if len(arquivosSelecionados) > 0 {
		utils.Logger("Iniciando o processo de importação dos "+fmt.Sprint(len(arquivosSelecionados))+" arquivos.", "info")

		for _, nomeArquivoDbc := range arquivosSelecionados {
			importarArquivo(nomeArquivoDbc)
		}

		//DeletarArquivosTemporarios()
		emailTerminoImportacao(arquivosSelecionados)
		utils.Logger("Processo de importação dos "+fmt.Sprint(len(arquivosSelecionados))+" arquivos finalizado. Arquivos persistidos: "+fmt.Sprint(qtdArquivoPersistido)+". Erros: "+fmt.Sprint(utils.ContErro)+".", "info")
		utils.Logger("Aguardando próximo agendamento.", "info")

	} else {
		utils.Logger("Não foram localizados arquivos para importação.", "info")
		utils.Logger("Aguardando próximo agendamento.", "info")
		emailSemArquivoNovo()
	}
}
