package services

import (
	"dados-apac/pkg/utils"
	"fmt"
	"os"
	"time"
)

// validacaoFinal ira validar os dados inseridos para um unico arquivo DBC
func validacaoFinal() {
	qtdRegistroBanco, err := contarRegistrosBanco()
	if err != nil {
		purgeInsertDbc(nomeArquivoDbc)
		return
	}

	if err := validarInsert(qtdRegistroBanco); err != nil {
		purgeInsertDbc(nomeArquivoDbc)
		return
	}

	if err := complementarAuditoria(qtdRegistroBanco); err != nil {
		purgeInsertDbc(nomeArquivoDbc)
		return
	}

	DeletarArquivosTemporarios()

	qtdArquivoPersistido++

	utils.Logger(nomeArquivoDbc+": ---------- Processo de importacao finalizado. ----------", "info")

}

// contarRegistrosBanco conta a quantidade de registros que foram gravados no banco para um unico arquivo DBC
func contarRegistrosBanco() (int, error) {
	qtdInsertArquivo, err := data.ConsultarQtdRegistroArquivo(nomeArquivoDbc)
	if err != nil {
		//se nao conseguiu verificar a quantidade de registros entao os registros inseridos serao excluidos
		utils.Logger(nomeArquivoDbc+": Erro ao consultar a qtd de registros no banco de dados. Todos os registros serao excluidos do banco de dados.", "error")
		return -1, err
	}

	return qtdInsertArquivo, nil
}

// validarInsert ira fazer as validacoes finais comparando os quantitativos contados durante a execucao do programa
func validarInsert(qtdInsertArquivo int) error {
	if qtdRegistroDbf != qtdRegistroLido {
		err := fmt.Errorf("%v: Divergencia. Arquivo foi recebido com %v registros e apenas %v foram lidos pelo programa. Todos os registros serao excluidos do banco de dados.", nomeArquivoDbc, qtdRegistroDbf, qtdRegistroLido)
		utils.Logger(err, "error")
		return err
	}

	if qtdInsertArquivo < qtdRegistroApac {
		err := fmt.Errorf("%v: Divergencia. Arquivo possui %v registros APAC e foram inseridos apenas %v no banco. Todos os registros serao excluidos do banco de dados.", nomeArquivoDbc, qtdRegistroApac, qtdInsertArquivo)
		utils.Logger(err, "error")
		return err
	} else {
		utils.Logger(nomeArquivoDbc+": Arquivo foi inserido no banco de dados com "+fmt.Sprint(qtdInsertArquivo)+" registros.", "info")
		return nil
	}
}

// complementarAuditoria atualiza a tabela auditoria com a quantidade de registros inseridos, quantidade de registros com a quantidade zerada e a data de finalizacao
func complementarAuditoria(qtdInsertArquivo int) error {
	dataFinal := time.Now()
	if err := data.UpdateAuditoria(nomeArquivoDbc, qtdInsertArquivo, qtdRegistroLido, qtdRegistroApac, dataFinal); err != nil {
		return err
	}

	utils.Logger(nomeArquivoDbc+": Tabela de auditoria foi atualizada.", "info")
	return nil
}

// purgeInsertDbc exclui os registros de um arquivo DBC em caso de erro.
// se a exclusao der erro o programa ira parar de funcionar pois o banco ficou inconsistente
func purgeInsertDbc(nomeArquivoDbc string) {
	if err := data.ExcluirMedicamento(nomeArquivoDbc); err != nil {
		utils.Logger(nomeArquivoDbc+": Erro ao excluir todos os registros do banco de dados para o arquivo. Banco de dados com registros inconsistentes.", "fatal")
	}

	if err := data.ExcluirAuditoria(nomeArquivoDbc); err != nil {
		utils.Logger(nomeArquivoDbc+": Erro ao excluir todos os registros do banco de dados para o arquivo. Banco de dados com registros inconsistentes.", "fatal")
	}

	utils.Logger(nomeArquivoDbc+": Os registros do arquivo foram excluidos do banco de dados.", "info")
}

// DeletarArquivosTemporarios ira apagar todos os arquivos das pastas dbcTemp, dbfTemp e csvTemp
func DeletarArquivosTemporarios() {
	if err := os.RemoveAll(dbcTempPath); err != nil {
		utils.Logger(nomeArquivoDbc+": Erro ao excluir os registros da pasta '"+dbcTempPath+"'.", "error")
	}
	os.MkdirAll(dbcTempPath, 0777)

	if err := os.RemoveAll(dbfTempPath); err != nil {
		utils.Logger(nomeArquivoDbc+": Erro ao excluir os registros da pasta '"+dbfTempPath+"'.", "error")
	}
	os.MkdirAll(dbfTempPath, 0777)

	if err := os.RemoveAll(csvTempPath); err != nil {
		utils.Logger(nomeArquivoDbc+": Erro ao excluir os registros da pasta '"+csvTempPath+"'.", "error")
	}
	os.MkdirAll(csvTempPath, 0777)
}
