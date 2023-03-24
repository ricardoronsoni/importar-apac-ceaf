package database

import (
	"dados-apac/pkg/utils"
	"time"
)

var arquivoPersistido bool

// ValidacaoFinalBanco é utilizada para representar a validacao final de consistencia do banco de dados
type ValidacaoFinalBanco struct {
	Arquivo       string
	RegistroBanco string
	CountRegistro string
}

// ConsultarArquivo verifica se um arquivo localizado no FTP ja foi persistido anteriormente
func (db Database) ConsultarArquivo(nomeArquivo string) bool {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
	}

	sqlStatement := "select count(*) from apac.auditoria where arquivo = $1 limit 1;"

	rows, err := db.SqlDb.QueryContext(dbContext, sqlStatement, nomeArquivo)
	if err != nil {
		utils.Logger(err, "error")
	}

	defer func() { _ = rows.Close() }()

	var qtdArquivo string
	for rows.Next() {
		err := rows.Scan(&qtdArquivo)
		if err != nil {
			utils.Logger(err, "error")
		}
	}

	if qtdArquivo == "0" {
		arquivoPersistido = false
	} else {
		arquivoPersistido = true
	}

	return arquivoPersistido
}

// InserirAuditoria insere no banco de dados os dados inciais de um arquivo que foi localizado no FTP
func (db Database) InserirAuditoria(registrosDbf int, nomeArquivo string, dataInicio time.Time) error {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return err
	}

	queryStatement := "insert into apac.auditoria(arquivo, registros_dbf, inicio_processo) values ($1, $2, $3)"

	stmt, err := db.SqlDb.PrepareContext(dbContext, queryStatement)
	if err != nil {
		utils.Logger(err, "error")
		return err
	}
	defer stmt.Close()

	if _, err = stmt.ExecContext(dbContext, nomeArquivo, registrosDbf, dataInicio.Format("2006/01/02 15:04:05.000")); err != nil {
		utils.Logger(err, "error")
		return err
	}

	utils.Logger(nomeArquivo+": Arquivo gravado na tabela de auditoria com os dados iniciais.", "info")

	return nil
}

// UpdateAuditoria ira atualizar a tabela de auditoria com a qtd de registros inseridos e a data final
func (db Database) UpdateAuditoria(nomeArquivo string, qtdInsertArquivo, qtdRegistroLido, qtdRegistroApac int, dataFim time.Time) error {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return err
	}

	queryStatement := "update apac.auditoria set registros_lidos = $1, registros_apac_medicamento = $2, registros_banco = $3, fim_processo = $4, data_alteracao = $5 where arquivo = $6"

	dataAtualizacao := time.Now()
	if _, err := db.SqlDb.ExecContext(dbContext, queryStatement, qtdRegistroLido, qtdRegistroApac, qtdInsertArquivo, dataFim.Format("2006/01/02 15:04:05.000"), dataAtualizacao.Format("2006/01/02 15:04:05.000"), nomeArquivo); err != nil {
		utils.Logger(err, "error")
		return err
	}

	return nil
}

// ExcluirAuditoria ira excluir todos os registros da tabela de auditoria
func (db Database) ExcluirAuditoria(nomeArquivo string) error {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return err
	}

	queryStatement := "delete from apac.auditoria where arquivo = $1"

	if _, err := db.SqlDb.ExecContext(dbContext, queryStatement, nomeArquivo); err != nil {
		utils.Logger(err, "error")
		return err
	}

	return nil
}

// InconsistenciaAuditoria verifica se existe algum arquivo que nao terminou o processamento
func (db Database) InconsistenciaAuditoria() ([]string, error) {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return []string{}, err
	}

	sqlStatement := "select arquivo from apac.auditoria where fim_processo is null;"

	rows, err := db.SqlDb.QueryContext(dbContext, sqlStatement)
	if err != nil {
		utils.Logger(err, "error")
		return []string{}, err
	}

	defer func() { _ = rows.Close() }()

	var arquivo = ""
	var arquivos = []string{}
	for rows.Next() {
		err := rows.Scan(&arquivo)
		if err != nil {
			utils.Logger(err, "error")
			return []string{}, err
		}
		arquivos = append(arquivos, arquivo)
	}

	return arquivos, nil
}

// InconsistenciaAuditoria verifica se existe algum arquivo que nao terminou o processamento
func (db Database) ValidacaoTotalFinal() ([]ValidacaoFinalBanco, error) {
	if err := db.SqlDb.PingContext(dbContext); err != nil {
		utils.Logger(err, "error")
		return []ValidacaoFinalBanco{}, err
	}

	sqlStatement := `select 
						a.arquivo,
						a.registros_banco,
						count(m.arquivo_origem)
					from apac.auditoria a 
					left join apac.medicamento m on a.arquivo = m.arquivo_origem 
					group by 
						a.arquivo,
						a.registros_banco
					having 	a.registros_banco != count(m.arquivo_origem)`

	rows, err := db.SqlDb.QueryContext(dbContext, sqlStatement)
	if err != nil {
		utils.Logger(err, "error")
		return []ValidacaoFinalBanco{}, err
	}

	defer func() { _ = rows.Close() }()

	arquivosInconsistente := []ValidacaoFinalBanco{}
	for rows.Next() {
		arquivoInconsistente := ValidacaoFinalBanco{}
		err := rows.Scan(&arquivoInconsistente.Arquivo, &arquivoInconsistente.RegistroBanco, &arquivoInconsistente.CountRegistro)
		if err != nil {
			utils.Logger(err, "error")
		}
		arquivosInconsistente = append(arquivosInconsistente, arquivoInconsistente)
	}

	return arquivosInconsistente, nil
}
