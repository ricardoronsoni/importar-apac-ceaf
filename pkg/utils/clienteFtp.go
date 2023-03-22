package utils

import (
	"time"

	"github.com/jlaffaye/ftp"
)

// FtpConnection cria a conexao com o FTP do Datasus
func FtpConnection() (*ftp.ServerConn, error) {
	//faz a conexao com o FTP
	ClientFtp, err := ftp.Dial(urlFtp, ftp.DialWithTimeout(time.Duration(timeoutSegundos)*time.Second), ftp.DialWithDisabledEPSV(true))
	if err != nil {
		Logger(err, "error")
		return nil, err
	}

	//realiza o login com usuario e senha padrao
	err = ClientFtp.Login("anonymous", "anonymous")
	if err != nil {
		Logger(err, "error")
		return nil, err
	}

	//muda do diretorio raiz para o do SIASUS
	err = ClientFtp.ChangeDir(diretorioArquivos)
	if err != nil {
		Logger(err.Error(), "error")
		return nil, err
	}

	return ClientFtp, nil
}
