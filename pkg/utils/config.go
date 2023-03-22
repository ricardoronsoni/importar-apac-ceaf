package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	//Rotina para consultar os arquivos no Datasus
	AgendamentoDatasus = "@midnight"
	urlFtp             = "ftp.datasus.gov.br:21"
	diretorioArquivos  = "/dissemin/publicos/SIASUS/200801_/Dados"
	timeoutSegundos    = 60

	//Range de data em que serao baixados os arquivos
	DataInicial = 2101 //formato AAMM
	DataFinal   = 2101 //formato AAMM

	//DbConnection
	DbHost     = "postgres-analise-dados"
	DbPort     = "5432"
	DbName     string
	DbUser     string
	DbPassword string

	//Email
	UserEmail              string
	PasswordEmail          string
	HostEmail              = "smtp-mail.outlook.com"
	PortEmail              = 587
	SubjectEmailNotFound   = "Monitor APAC não localizou novos arquivos"
	SubjectEmailFoundFiles = "Monitor APAC localizou novos arquivos"
	SubjectFinishImport    = "Monitor APC finalizou a importação dos arquivos"
	DestinedEmail          = []string{"teste123@gmail.com"}
)

// CarregarEnv carrega as variaveis de ambiente do arquivo .env
func CarregarEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erro ao carregar o arquivo .env.")
	}

	DbName = os.Getenv("DB_NAME")
	DbUser = os.Getenv("DB_USER")
	DbPassword = os.Getenv("DB_PASSWORD")
	UserEmail = os.Getenv("EMAIL_USER")
	PasswordEmail = os.Getenv("EMAIL_PASSWORD")
}
