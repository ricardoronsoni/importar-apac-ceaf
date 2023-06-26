package services

import (
	"bytes"
	"dados-apac/pkg/utils"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// envioLooker faz a etração dos dados inseridos pela rotina em CSV e ransmite os dados
func envioLooker() {
	utils.Logger("Obtendo as competências de dispensação inseridas.", "info")

	competenciasInseridas, err := data.ComeptenciasInseridas(inicioProcesso)
	if err != nil {
		return
	}

	if len(competenciasInseridas) == 0 {
		return
	}

	utils.Logger("Gerando arquivo CSV com os dados inseridos.", "info")

	rows, err := data.ExtracaoLooker(competenciasInseridas)
	if err != nil {
		return
	}

	if gerarCsv(rows); err != nil {
		return
	}

	if transmitirLooker(competenciasInseridas); err != nil {
		return
	}

	if err := os.RemoveAll(csvTempPath); err != nil {
		utils.Logger("extracaoFinal.csv: Erro ao excluir o arquivo'.", "error")
	}

}

// gerarCsv gera o arquivo CSV a partir do resultado obtido do banco de dadoss
func gerarCsv(rows *sql.Rows) error {
	defer rows.Close()

	file, err := os.Create(csvTempPath + "extracaoFinal.csv")
	if err != nil {
		utils.Logger(err, "error")
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	//criar o acbeçalho
	err = writer.Write([]string{"competencia_dispensacao", "sigla_uf_dispensacao", "regiao_uf", "procedimento", "descricao_procedimento", "cid",
		"faixa_etaria_paciente", "sexo_paciente", "quantidade_aprovada", "valor_aprovado", "total_pacientes"})
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	for rows.Next() {
		var atributo1, atributo2, atributo3, atributo4, atributo5, atributo6, atributo7, atributo8, atributo9,
			metrica1, metrica2, metrica3 string

		err = rows.Scan(&atributo1, &atributo2, &atributo3, &atributo4, &atributo5, &atributo6, &atributo7, &atributo8, &atributo9, &metrica1, &metrica2, &metrica3)
		if err != nil {
			utils.Logger(err, "error")
			return err
		}

		err = writer.Write([]string{atributo1, atributo2, atributo3, atributo4, atributo5, atributo6, atributo7, atributo8, atributo9, metrica1, metrica2, metrica3})
		if err != nil {
			utils.Logger(err, "error")
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		utils.Logger(err, "error")
		return err
	}

	utils.Logger("extracaoFinal.csv: Arquivo gerado com sucesso.", "info")

	return nil
}

func transmitirLooker(competenciasInseridas []string) {
	utils.Logger("extracaoFinal.csv: Transmitindo os dados para o servidor cloud.", "info")

	filePath := csvTempPath + "extracaoFinal.csv"

	competenciasString := strings.Join(competenciasInseridas, ",")
	url := utils.UrlCloudLooker + "?competencia=" + competenciasString

	file, err := os.Open(filePath)
	if err != nil {
		utils.Logger(err, "error")
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		utils.Logger(err, "error")
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		utils.Logger(err, "error")
		return
	}

	err = writer.Close()
	if err != nil {
		utils.Logger(err, "error")
		return
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		utils.Logger(err, "error")
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("API-Key", utils.ApiKeyCloud)

	// Cria um cliente HTTP personalizado com o tempo de timeout de 10 minutos
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		utils.Logger(err, "error")
		return
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			utils.Logger(err, "error")
			return
		}
	}
	defer resp.Body.Close()

	//Lê o corpo da resposta
	bodyJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.Logger(err, "error")
		return
	}

	// Verifica o status da resposta
	if resp.StatusCode != http.StatusOK {
		utils.Logger(fmt.Sprintf("Erro ao gravar os dados no servidor Clud: Status: %s - %s", resp.Status, bodyJson), "error")
		return
	}

	utils.Logger("extracaoFinal.csv: Dados gravados no servidor cloud.", "info")

}
