package services

import (
	"dados-apac/pkg/database"
	"dados-apac/pkg/utils"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/LindsayBradford/go-dbf/godbf"
	"github.com/tadvi/dbf"
)

var (
	listaArquivosDbf []string
	qtdRegistroDbf   int
	qtdRegistroLido  int
	qtdRegistroApac  int
)

// converterArquivos ira converter os arquivos DBC em DBF e CSV
func converterArquivos() error {
	if err := verificarDuplicidade(); err != nil {
		return err
	}

	if err := dbcParaDbf(); err != nil {
		return err
	}

	removeDbc()

	if err := contarRegistrosDbf(); err != nil {
		return err
	}

	data.InserirAuditoria(qtdRegistroDbf, nomeArquivoDbc, dataInicio)

	if err := dbfPaParaCsv(); err != nil {
		return err
	}

	if err := data.TruncateTempMedicamento(); err != nil {
		return err
	}

	if err := data.InserirTempMedicamento(nomeArquivoDbc, csvTempPath+strings.TrimSuffix(nomeArquivoDbc, ".dbc")+".csv"); err != nil {
		return err
	}

	removeCsv()

	if err := lerArquivosAm(); err != nil {
		return err
	}

	removeDbf()

	if err := data.LocalizarRegistroPaSemComplemento(nomeArquivoDbc); err != nil {
		return err
	}

	if err := data.CopiarTabelaDefinitiva(nomeArquivoDbc); err != nil {
		return err
	}

	validacaoFinal()

	return nil
}

// verificarDuplicidade ira verificar se existe algum registro do arquivo a ser importado na tabela definitiva. Se houve, os registros serao excluidos
func verificarDuplicidade() error {
	qtdRegistroPersistido, err := data.ConsultarDuplicidade(nomeArquivoDbc)
	if err != nil {
		return err
	}

	if qtdRegistroPersistido > 0 {
		err := fmt.Errorf(nomeArquivoDbc + ": Arquivo ja possui registros gravados na tabela definitiva de medicamento. Para evitar duplicidade os mesmos serao deletados.")
		utils.Logger(err, "info")
		purgeInsertDbc(nomeArquivoDbc)
	}

	return nil
}

// dbcParaDbf faz a conversao do arquivo dbc para dbf
func dbcParaDbf() error {
	listaArquivosDbf = []string{}
	for _, nomeArquivo := range listaArquivosDownload {
		arquivoDbf := strings.TrimSuffix(nomeArquivo, ".dbc") + ".dbf"
		listaArquivosDbf = append(listaArquivosDbf, arquivoDbf)

		//usa comando shell para acionar o programa em C na pasta ./dbc/dbcParaDbf. Readme tem informacoes desse programa
		cmd := exec.Command("/bin/sh", "-c", "cd "+dbcParaDbfPath+"; ./blast-dbf ../dbcTemp/"+nomeArquivo+" ../dbfTemp/"+arquivoDbf)
		err := cmd.Run()
		if err != nil {
			utils.Logger(err, "error")
			return err
		}

		utils.Logger(nomeArquivo+": Arquivo convertido para DBF.", "info")
	}

	return nil
}

// removeDbc exclui o arquivo DBC que ja transformado em DBF
func removeDbc() {
	for _, nomeArquivo := range listaArquivosDownload {
		err := os.Remove(dbcTempPath + nomeArquivo)
		if err != nil {
			utils.Logger(err, "error")
		}

		utils.Logger(nomeArquivo+": Arquivo excluido.", "info")
	}
}

// contarRegistrosDbf conta a quantidade de registros totais que o arquivo PA, contando dados que nao APAC de medicamentos
func contarRegistrosDbf() error {
	var (
		dbfArquivo *godbf.DbfTable
		err        error
	)

	qtdRegistroDbf = 0
	for _, arquivoDbf := range listaArquivosDbf {
		if arquivoDbf[0:2] == "PA" {
			dbfArquivo, err = godbf.NewFromFile(dbfTempPath+arquivoDbf, "UTF8")
			if err != nil {
				utils.Logger(err, "error")
				return err
			}
			qtdRegistroDbf += dbfArquivo.NumberOfRecords()
		}
		dbfArquivo = nil
	}

	return nil
}

// removeDbf exclui o arquivo DBC que ja foi transformado em DBF
func removeDbf() {
	for _, nomeArquivo := range listaArquivosDbf {
		err := os.Remove(dbfTempPath + nomeArquivo)
		if err != nil {
			utils.Logger(err, "error")
		}

		utils.Logger(nomeArquivo+": Arquivo excluido.", "info")
	}
}

// dbfAmPaParaCsv abre o arquivo DBF do tipo PA e grava os registros no arquivo CSV
func dbfPaParaCsv() error {
	qtdRegistroLido = 0
	qtdRegistroApac = 0

	csvFile, err := os.Create(csvTempPath + strings.TrimSuffix(nomeArquivoDbc, ".dbc") + ".csv")
	if err != nil {
		utils.Logger(err, "error")
		return err
	}
	defer csvFile.Close()

	for _, arquivoDbf := range listaArquivosDbf {
		if arquivoDbf[0:2] == "PA" {

			db, err := dbf.LoadFile(dbfTempPath + arquivoDbf)
			if err != nil {
				utils.Logger(err, "error")
				return err
			}

			iter := db.NewIterator()
			for iter.Next() {
				registroPa := iter.Row()
				qtdRegistroLido++

				//para pegar apenas os registros com qtd aprovada e do tipo APAC
				if registroPa[15][0:4] == "0604" { // && (registroPa[19] == "P" || registroPa[19] == "S") {
					qtdRegistroApac++

					registroApacTratados, err := tratarDadosPa(registroPa, nomeArquivoDbc)
					if err == nil {
						err = inserirRegistroCsv(registroApacTratados, csvFile)
						if err != nil {
							break
						}
					} else {
						break
					}
				}
			}
			//liberar o que esta gravado em memoria
			db = nil
		}
	}

	utils.Logger(nomeArquivoDbc+": Arquivo possui "+fmt.Sprint(qtdRegistroApac)+" registros.", "info")

	return nil
}

// tratarDadosPa ira tratar os dados recebidos no arquivo do Datasus
func tratarDadosPa(registroApac []string, nomeArquivo string) (database.DadosApac, error) {
	var dados = database.DadosApac{}
	var err error

	dados.CompetenciaDispensacaoString = registroApac[14]
	if dados.CompetenciaDispensacao, err = time.Parse("200601", registroApac[14]); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if dados.CompetenciaProcessamento, err = time.Parse("200601", registroApac[13]); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if dados.CodigoUfDispensacao, err = strconv.Atoi(registroApac[1][0:2]); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if dados.SiglaUfDispensacao, err = pegarUf(dados.CodigoUfDispensacao); err != nil {
		return database.DadosApac{}, err
	}
	if dados.MunicipioDispensacao, err = strconv.Atoi(registroApac[3]); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	dados.CnesDispensacao = registroApac[0]
	dados.Apac = registroApac[20]
	dados.Procedimento = registroApac[15]
	dados.Cid = registroApac[29]
	if dados.QtdApresentada, err = strconv.Atoi(registroApac[40]); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if dados.ValorApresentado, err = strconv.ParseFloat(registroApac[42], 64); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if dados.QtdAprovada, err = strconv.Atoi(registroApac[41]); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if dados.ValorAprovado, err = strconv.ParseFloat(registroApac[43], 64); err != nil {
		utils.Logger(err, "error")
		return database.DadosApac{}, err
	}
	if registroApac[33] == "" {
		dados.IdadePaciente = -1
	} else {
		if dados.IdadePaciente, err = strconv.Atoi(registroApac[33]); err != nil {
			utils.Logger(err, "error")
			return database.DadosApac{}, err
		}
	}
	dados.SexoPaciente = registroApac[37]
	dados.RacaCorPaciente = registroApac[38]
	dados.EtniaPaciente = registroApac[53]
	if registroApac[39] == "" {
		dados.MunicipioResidenciaPaciente = -1
	} else {
		if dados.MunicipioResidenciaPaciente, err = strconv.Atoi(registroApac[39]); err != nil {
			utils.Logger(err, "error")
			return database.DadosApac{}, err
		}
	}
	dados.ObitoPaciente = registroApac[24]
	dados.ArquivoOrigem = nomeArquivo

	return dados, nil

}

// tratarDadosAm ira tratar os dados do arquivo AM para suplementar os registros PA
func tratarDadosAm(registroApac []string) (database.DadosApac, error) {
	var dados = database.DadosApac{}
	var err error

	dados.Apac = registroApac[4]
	dados.CnsPaciente = strings.ReplaceAll(fmt.Sprintf("% x", registroApac[14]), " ", "")
	if registroApac[45] == "" {
		dados.PesoPaciente = -1
	} else {
		if dados.PesoPaciente, err = strconv.Atoi(registroApac[45]); err != nil {
			utils.Logger(err, "error")
			return database.DadosApac{}, err
		}
	}
	if registroApac[46] == "" {
		dados.AlturaPaciente = -1
	} else {
		if dados.AlturaPaciente, err = strconv.Atoi(registroApac[46]); err != nil {
			utils.Logger(err, "error")
			return database.DadosApac{}, err
		}
	}
	dados.NacionalidadePaciente = registroApac[20]
	dados.CnesSolicitante = registroApac[38]
	if registroApac[39] == "" {
		dados.DataSolicitacao = time.Time{}
	} else {
		if dados.DataSolicitacao, err = time.Parse("20060102", registroApac[39]); err != nil {
			utils.Logger(err, "error")
			return database.DadosApac{}, err
		}
	}
	if registroApac[40] == "" {
		dados.DataAutorizacao = time.Time{}
	} else {
		if dados.DataAutorizacao, err = time.Parse("20060102", registroApac[40]); err != nil {
			utils.Logger(err, "error")
			return database.DadosApac{}, err
		}
	}

	return dados, nil
}

// lerArquivosAm ira abrir o arquivo DBF do tipo AM, limpar os dados e fazer update na tabela temporaria do banco de dados com os dados do arquivo PA
func lerArquivosAm() error {
	db, err := dbf.LoadFile(dbfTempPath + strings.TrimSuffix(nomeArquivoDbc, ".dbc") + ".dbf")
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	utils.Logger(nomeArquivoDbc+": Iniciando a atualizacao da tabela temporaria com os dados do arquivo AM.", "info")

	iter := db.NewIterator()
	for iter.Next() {
		registroAm := iter.Row()

		amTratado, err := tratarDadosAm(registroAm)
		if err != nil {
			return err
		}

		err = data.UpdateTempMedicamento(amTratado, nomeArquivoDbc)
		if err != nil {
			return err
		}
	}
	utils.Logger(nomeArquivoDbc+": Tabela temporaria foi atualizada com os dados do arquivo AM.", "info")

	return nil
}

// inserirRegistroCsv escreve um registro no arquivo CSV
func inserirRegistroCsv(registro database.DadosApac, csvFile *os.File) error {
	w := csv.NewWriter(csvFile)
	defer w.Flush()

	if err := w.Write([]string{
		registro.CompetenciaDispensacao.Format("2006-01-02"),
		registro.CompetenciaProcessamento.Format("2006-01-02"),
		strconv.Itoa(registro.CodigoUfDispensacao),
		registro.SiglaUfDispensacao,
		strconv.Itoa(registro.MunicipioDispensacao),
		registro.CnesDispensacao,
		registro.Apac,
		registro.Procedimento,
		registro.Cid,
		strconv.Itoa(registro.QtdApresentada),
		strconv.Itoa(registro.QtdAprovada),
		fmt.Sprintf("%.2f", registro.ValorApresentado),
		fmt.Sprintf("%.2f", registro.ValorAprovado),
		strconv.Itoa(registro.IdadePaciente),
		registro.SexoPaciente,
		registro.RacaCorPaciente,
		registro.EtniaPaciente,
		strconv.Itoa(registro.MunicipioResidenciaPaciente),
		registro.ObitoPaciente,
		registro.ArquivoOrigem,
	},
	); err != nil {
		utils.Logger(err, "error")
		return err
	}

	return nil
}

// removeCsv exclui o arquivo DBC que ja foi transformado em DBF
func removeCsv() {
	arquivoCsv := strings.TrimSuffix(nomeArquivoDbc, ".dbc") + ".csv"
	err := os.Remove(csvTempPath + arquivoCsv)
	if err != nil {
		utils.Logger(err, "error")
	}

	utils.Logger(arquivoCsv+": Arquivo excluido.", "info")
}

// pegarUf retorna a sigla da UF para um codigo ibge informado
func pegarUf(codigoIbge int) (string, error) {
	switch codigoIbge {
	case 11:
		return "RO", nil
	case 12:
		return "AC", nil
	case 13:
		return "AM", nil
	case 14:
		return "RR", nil
	case 15:
		return "PA", nil
	case 16:
		return "AP", nil
	case 17:
		return "TO", nil
	case 21:
		return "MA", nil
	case 22:
		return "PI", nil
	case 23:
		return "CE", nil
	case 24:
		return "RN", nil
	case 25:
		return "PB", nil
	case 26:
		return "PE", nil
	case 27:
		return "AL", nil
	case 28:
		return "SE", nil
	case 29:
		return "BA", nil
	case 31:
		return "MG", nil
	case 32:
		return "ES", nil
	case 33:
		return "RJ", nil
	case 35:
		return "SP", nil
	case 41:
		return "PR", nil
	case 42:
		return "SC", nil
	case 43:
		return "RS", nil
	case 50:
		return "MS", nil
	case 51:
		return "MT", nil
	case 52:
		return "GO", nil
	case 53:
		return "DF", nil
	default:
		err := fmt.Errorf(nomeArquivoDbc + ": Codigo UF " + fmt.Sprint(codigoIbge) + " nao localizado.")
		utils.Logger(err, "error")
		return "", err
	}
}
