package services

import (
	"dados-apac/pkg/utils"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	bodySemArquivosNovos    string
	bodyArquivosLocalizados string
	bodyTerminoImportacao   string
)

// emailSemArquivosNovos envia email quando nao foram localizados novos arquivos
func emailSemArquivoNovo() {
	templateSemArquivoNovo()
	clientEmail("semArquivoNovo")
}

// emailArquivosLocalizados envia email quando foi localizado novos arquivos
func emailArquivosLocalizados(arquivos []string) {
	templateArquivosLocalizados(arquivos)
	clientEmail("arquivosLocalizados")
}

// emailTerminoImportacao envia email quando o processo de improtacao for finalizado
func emailTerminoImportacao(arquivos []string) {
	templateTerminoImportacao(arquivos)
	clientEmail("terminoImportacao")
}

// templateSemArquivosNovos monta o body do email que informa que nao foram localizados novos arquivos
func templateSemArquivoNovo() {
	greeting := getGreeting()

	bodySemArquivosNovos = fmt.Sprintf(
		`%v,<br>
		<br>
		Foi verificado às %v que o FTP do Datasus não possui novos arquivos do SIA/SUS para importação.<br>
		<br>
		Este é um email automático do sistema. Não responda o mesmo.<br>
		<br>
		Att,<br>
		DAF/SCTIE/MS`,
		greeting,
		time.Now().Format("15h04m05s (02/01/2006)"),
	)
}

// templateArquivosLocalizados gera o body do email com registros localizados
func templateArquivosLocalizados(arquivos []string) {

	greeting := getGreeting()

	header := fmt.Sprintf(
		`%v,<br>
		<br>
		Foi verificado às %v que %v novos arquivos no FTP do Datasus estão disponíveis para importação, conforme indicado abaixo:<br>
		<br>
		<strong><u>Arquivos:</u></strong><br>`,
		greeting, inicioProcesso.Format("15h04m05s (02/01/2006)"), len(arquivos),
	)

	var errorList string
	for _, arquivo := range arquivos {
		errorList += fmt.Sprintf("     - %v<br>", arquivo)
	}

	footer := `
		<br>
		Um email automático será enviado assim que o processo de importação de todos os arquivos for finalizado.<br>
		<br>
		Este é um email automático do sistema. Não responda o mesmo.<br>
		<br>
		Att,<br>
		DAF/SCTIE/MS`

	bodyArquivosLocalizados = header + errorList + footer
}

// templateTerminoImportacao monta o body do email que informa que o processo de importacao dos arquivos finalizou
func templateTerminoImportacao(arquivos []string) {
	greeting := getGreeting()

	bodyTerminoImportacao = fmt.Sprintf(
		`%v,<br>
		<br>
		O processo de importação iniciado %v contendo %v arquivos foi finalizado. No total %v arquivos foram importados com sucesso.<br>
		<br>
		Durante o processo de importação foram encontrados `+fmt.Sprint(utils.ContErro)+` erros. Os erros ficam disponíveis para consulta no arquivo '`+utils.PathErro+`'.<br>
		<br>
		Este é um email automático do sistema. Não responda o mesmo.<br>
		<br>
		Att,<br>
		DAF/SCTIE/MS`,
		greeting,
		inicioProcesso.Format("15h04m05s (02/01/2006)"),
		len(arquivos),
		qtdArquivoPersistido,
	)
}

// clientEmail envia os emails
func clientEmail(emailType string) {
	m := gomail.NewMessage()

	for _, email := range utils.DestinedEmail {
		m.SetHeader("From", utils.UserEmail)
		m.SetHeader("To", email)

		if emailType == "semArquivoNovo" {
			m.SetHeader("Subject", utils.SubjectEmailNotFound)
			m.SetBody("text/html", bodySemArquivosNovos)
		} else if emailType == "arquivosLocalizados" {
			m.SetHeader("Subject", utils.SubjectEmailFoundFiles)
			m.SetBody("text/html", bodyArquivosLocalizados)
		} else if emailType == "terminoImportacao" {
			m.SetHeader("Subject", utils.SubjectFinishImport)
			m.SetBody("text/html", bodyTerminoImportacao)
		}

		d := gomail.NewDialer(utils.HostEmail, utils.PortEmail, utils.UserEmail, utils.PasswordEmail)

		if err := d.DialAndSend(m); err != nil {
			utils.Logger(err, "error")
		}
	}
}

// getGreeting determines the greeting at the beginning of the email with errors
func getGreeting() string {
	var greeting string

	hour, err := strconv.Atoi(time.Unix(time.Now().UnixMilli()/1000, 0).Format("15"))
	if err != nil {
		utils.Logger(err, "error")
	}

	if hour >= 5 && hour < 12 {
		greeting = "Bom dia"
	} else if hour >= 12 && hour < 18 {
		greeting = "Boa tarde"
	} else if hour >= 18 || hour < 5 {
		greeting = "Boa noite"
	}

	return greeting
}
