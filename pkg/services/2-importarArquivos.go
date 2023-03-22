package services

import (
	"dados-apac/pkg/utils"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	listaArquivosDownload []string
	nomeArquivoDbc        string
	dataInicio            time.Time
)

// importarArquivoAm ira iniciar o processo de persistencia de um arquivo dos N arquivos localizados para importacao no FTP do Datasus
func importarArquivo(arquivoDbc string) {
	DeletarArquivosTemporarios()

	nomeArquivoDbc = arquivoDbc

	if err := localizarInconsistenciasBanco(); err != nil {
		return
	}

	utils.Logger(nomeArquivoDbc+": Processo de importação do arquivo iniciado.", "info")
	dataInicio = time.Now()

	listaArquivosDownload = []string{nomeArquivoDbc} //a posicao 0 dessa variavel sempre sera do arquivo AM. As outras posicoes sao dos arquivos PA

	if err := listarArquivosPa(); err != nil {
		return
	}

	if err := downloadFtp(); err != nil {
		return
	}

	if err := compararTamanhos(); err != nil {
		return
	}

	if err := converterArquivos(); err != nil {
		return
	}

}

// listarArquivosPa ira listar as variacoes (final a, b, c, ...) do arquivo PA disponiveis para download no FTP do Datasus
func listarArquivosPa() error {
	nomeArquivoPa := "PA" + nomeArquivoDbc[2:]
	for _, arquivoFtp := range todosArquivos {
		if arquivoFtp.Name[0:8] == nomeArquivoPa[0:8] {
			listaArquivosDownload = append(listaArquivosDownload, arquivoFtp.Name)
			utils.Logger(arquivoFtp.Name+": Processo de importação do arquivo iniciado.", "info")
		}
	}

	if len(listaArquivosDownload) <= 1 {
		utils.Logger(listaArquivosDownload[0]+": Não foi localizado o arquivo "+nomeArquivoPa+". Processo de importação cancelado.", "error")
	}
	return nil
}

// downloadFtp vai fazer o download do arquivo dbc do FTP do Datasus
func downloadFtp() error {
	for _, nomeArquivo := range listaArquivosDownload {
		var err error
		//faz a conexao para baixar cada arquivo para evitar o erro 229 do FTP
		clientFtp, err = utils.FtpConnection()
		if err != nil {
			return err
		}

		arquivoDbc, err := clientFtp.Retr(nomeArquivo)
		if err != nil {
			utils.Logger(err, "error")
			return err
		}

		defer arquivoDbc.Close()

		outFile, err := os.Create(dbcTempPath + nomeArquivo)
		if err != nil {
			utils.Logger(err, "error")
			return err
		}

		defer outFile.Close()

		_, err = io.Copy(outFile, arquivoDbc)
		if err != nil {
			utils.Logger(err, "error")
			return err
		}

		utils.Logger(nomeArquivo+": Arquivo baixado com sucesso do FTP do Datasus.", "info")
	}
	return nil
}

// compararTamanhos compara o tamanho do arquivo DBC no FTP e no disco do servidor local
func compararTamanhos() error {
	for _, nomeArquivo := range listaArquivosDownload {
		tamanhoFtp, err := clientFtp.FileSize(nomeArquivo)
		if err != nil {
			utils.Logger(err, "fatal")
			return err
		}

		tamanhoDisco, err := os.Stat(dbcTempPath + nomeArquivo)
		if err != nil {
			utils.Logger(err, "fatal")
			return err
		}

		if tamanhoFtp == tamanhoDisco.Size() {
			utils.Logger(nomeArquivo+": Arquivo salvo em disco possui o mesmo tamanho do FTP: "+fmt.Sprintf("%.6f", float64(tamanhoFtp)/float64(1000000))+"Mb", "info")
		} else {
			err := fmt.Errorf("%v: Arquivo salvo em disco possui %vb e no FTP possui %vb.", nomeArquivo, tamanhoDisco.Size(), tamanhoFtp)
			utils.Logger(err, "fatal")
			return err
		}
	}
	return nil
}
