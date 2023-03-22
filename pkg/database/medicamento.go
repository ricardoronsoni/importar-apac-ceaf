package database

import (
	"dados-apac/pkg/utils"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// DadosApac representa os campos do banco
type DadosApac struct {
	CompetenciaDispensacao       time.Time
	CompetenciaDispensacaoString string //competencia de dispensacao esta duplicada em formatos distintos para ganhar performace na compelementacao dos dados AM
	CompetenciaProcessamento     time.Time
	CodigoUfDispensacao          int
	SiglaUfDispensacao           string
	MunicipioDispensacao         int
	CnesDispensacao              string
	Apac                         string
	Procedimento                 string
	Cid                          string
	QtdApresentada               int
	ValorApresentado             float64
	QtdAprovada                  int
	ValorAprovado                float64
	CnsPaciente                  string
	IdadePaciente                int
	SexoPaciente                 string
	RacaCorPaciente              string
	EtniaPaciente                string
	PesoPaciente                 int
	AlturaPaciente               int
	MunicipioResidenciaPaciente  int
	NacionalidadePaciente        string
	ObitoPaciente                string
	CnesSolicitante              string
	DataSolicitacao              time.Time
	DataAutorizacao              time.Time
	ArquivoOrigem                string
}

// ConsultarDuplicidade verifica se um arquivo DBC possui dados na tabela definitiva
func (db Database) ConsultarDuplicidade(nomeArquivo string) (int, error) {
	err := db.SqlDb.PingContext(dbContext)
	if err != nil {
		utils.Logger(err, "error")
		return -1, err
	}

	sqlStatement := "select count(*) from apac.medicamento where arquivo_origem = $1 limit 1;"

	rows, err := db.SqlDb.QueryContext(dbContext, sqlStatement, nomeArquivo)
	if err != nil {
		utils.Logger(err, "error")
		return -1, err
	}

	defer func() { _ = rows.Close() }()

	var qtdRegistro string
	for rows.Next() {
		err := rows.Scan(&qtdRegistro)
		if err != nil {
			utils.Logger(err, "error")
			return -1, err
		}
	}

	var qtdRegistroInt int
	if qtdRegistroInt, err = strconv.Atoi(qtdRegistro); err != nil {
		utils.Logger(err, "error")
		return -1, err
	}

	return qtdRegistroInt, nil
}

// InserirTempMedicamento insere o arquivo CSV com os dados do arquivo PA na tabela temporaria de medicamento
func (db Database) InserirTempMedicamento(nomeArquivoDbc, csvFile string) error {
	utils.Logger(nomeArquivoDbc+": Iniciando a gravacao do arquivo PA na tabela temporaria do banco de dados.", "info")

	txn, err := db.SqlDb.Begin()
	if err != nil {
		utils.Logger(err, "error")
		return err
	}
	stmt, err := txn.Prepare(pq.CopyInSchema("apac", "temp_medicamento", "competencia_dispensacao", "competencia_processamento", "codigo_uf_dispensacao", "sigla_uf_dispensacao",
		"codigo_municipio_dispensacao", "cnes_dispensacao", "apac", "procedimento", "cid", "quantidade_apresentada", "quantidade_aprovada", "valor_apresentado", "valor_aprovado",
		"idade_paciente", "sexo_paciente", "raca_cor_paciente", "etnia_paciente", "municipio_residencia_paciente", "obito_paciente", "arquivo_origem"))
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	file, err := os.Open(csvFile)
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	csvReader := csv.NewReader(file)
	defer file.Close()
	for {
		registro, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			utils.Logger(err, "error")
			return err
		}

		_, err = stmt.Exec(registro[0], registro[1], registro[2], registro[3], registro[4], registro[5], registro[6], registro[7], registro[8], registro[9],
			registro[10], registro[11], registro[12], registro[13], registro[14], registro[15], registro[16], registro[17], registro[18], registro[19])
		if err != nil {
			utils.Logger(err, "error")
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		utils.Logger(err, "error")
		return err
	}
	err = stmt.Close()
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	err = txn.Commit()
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	utils.Logger(nomeArquivoDbc+": Gravacao dos registros na tabela temporaria finalizada.", "info")

	return nil
}

// UpdateTempMedicamento atualiza os registros de APAC do arquivo PA que estao na tabela temporaria do banco de dados
func (db Database) UpdateTempMedicamento(registro DadosApac, nomeArquivoDbc string) error {
	// if err := db.SqlDb.PingContext(dbContext); err != nil {
	// 	utils.Logger(err, "error")
	// 	return err
	// }

	queryStatement := `update apac.temp_medicamento set 
		cns_paciente = $1, peso_paciente = $2, altura_paciente = $3, nacionalidade_paciente = $4, cnes_solicitante = $5, data_solicitacao = $6, data_autorizacao = $7, data_alteracao = $8
		where apac = $9`

	var (
		dataAtualizacao = time.Now()
		res             sql.Result
		err             error
	)

	if res, err = db.SqlDb.ExecContext(dbContext, queryStatement, registro.CnsPaciente, registro.PesoPaciente, registro.AlturaPaciente,
		registro.NacionalidadePaciente, registro.CnesSolicitante, registro.DataSolicitacao, registro.DataAutorizacao,
		dataAtualizacao.Format("2006/01/02 15:04:05.000"), registro.Apac); err != nil {
		utils.Logger(err, "error")
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	if rowsAffected == 0 {
		err := fmt.Errorf(nomeArquivoDbc+": A APAC PA n. %v nao foi localizada para complementar com os dados do arquivo AM.", registro.Apac)
		utils.Logger(err, "error")
		return err
	}

	return nil
}

// LocalizarRegistroPaSemComplemento ira pesquisar os registros PA que ficaram sem complemento com o arquivo AM
// e entao ira pesquisar os dados faltantes no historico dos registros para complementar
func (db Database) LocalizarRegistroPaSemComplemento(nomeArquivoDbc string) error {
	queryStatement := `update apac.temp_medicamento set 
			cns_paciente = ac.cns_paciente, peso_paciente = ac.peso_paciente, altura_paciente = ac.altura_paciente, nacionalidade_paciente = ac.nacionalidade_paciente,
			cnes_solicitante = ac.cnes_solicitante, data_solicitacao = ac.data_solicitacao, data_autorizacao = ac.data_autorizacao, data_alteracao = $1
		from (
			select apac, cns_paciente, peso_paciente, altura_paciente, nacionalidade_paciente, cnes_solicitante, data_solicitacao, data_autorizacao
			from apac.medicamento 
			where substring(arquivo_origem, 3, 2) = $2
		) as ac
		where 
			temp_medicamento.apac = ac.apac
			and temp_medicamento.cns_paciente is null`

	var (
		dataAtualizacao = time.Now()
		nomeArquivo     = nomeArquivoDbc[2:4]
		err             error
	)

	if _, err = db.SqlDb.ExecContext(dbContext, queryStatement, dataAtualizacao, nomeArquivo); err != nil {
		utils.Logger(err, "error")
		return err
	}

	return nil
}

// TruncateTempMedicamento realiza o truncate na tabela temporaria de medicamentos
func (db Database) TruncateTempMedicamento() error {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return err
	}

	queryStatement := "truncate apac.temp_medicamento"

	if _, err := db.SqlDb.ExecContext(dbContext, queryStatement); err != nil {
		utils.Logger(err, "error")
		return err
	}

	return nil
}

// ConsultarQtdRegistroArquivo consulta a quantidade de registros persistidos para cada dbc do FTP
func (db Database) ConsultarQtdRegistroArquivo(nomeArquivo string) (int, error) {
	err := db.SqlDb.PingContext(dbContext)
	if err != nil {
		utils.Logger(err, "error")
		return -1, err
	}

	sqlStatement := "select count(*) from apac.medicamento where arquivo_origem = $1;"

	rows, err := db.SqlDb.QueryContext(dbContext, sqlStatement, nomeArquivo)
	if err != nil {
		utils.Logger(err, "error")
		return -1, err
	}

	defer func() { _ = rows.Close() }()

	var qtdRegistro string
	for rows.Next() {
		err := rows.Scan(&qtdRegistro)
		if err != nil {
			utils.Logger(err, "error")
			return -1, err
		}
	}

	var qtdRegistroInt int
	if qtdRegistroInt, err = strconv.Atoi(qtdRegistro); err != nil {
		utils.Logger(err, "error")
		return -1, err
	}

	return qtdRegistroInt, nil
}

// CopiarTabelaDefinitiva copia os dados da tabela temporaria para a tabela definitiva
func (db Database) CopiarTabelaDefinitiva(nomeArquivoDbc string) error {
	utils.Logger(nomeArquivoDbc+": Iniciando a gravacao na tabela definitiva de medicamento.", "info")

	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return err
	}

	queryStatement := "insert into apac.medicamento select * from apac.temp_medicamento"

	if _, err := db.SqlDb.ExecContext(dbContext, queryStatement); err != nil {
		utils.Logger(err, "error")
		return err
	}

	utils.Logger(nomeArquivoDbc+": Gravacao na tabela definitiva de medicamento finalizada.", "info")

	return nil
}

// ExcluirAuditoria ira excluir todos os registros da tabela de medicamento
func (db Database) ExcluirMedicamento(nomeArquivo string) error {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return err
	}

	queryStatement := "delete from apac.medicamento where arquivo_origem = $1"

	if _, err := db.SqlDb.ExecContext(dbContext, queryStatement, nomeArquivo); err != nil {
		utils.Logger(err, "error")
		return err
	}

	return nil
}
