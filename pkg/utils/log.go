package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	ContErro int
	pathLog  = "./log/" + time.Now().Format("2006-01-02") + ".log"
	PathErro = "./log/Erros " + time.Now().Format("2006-01-02") + ".txt"
)

const ()

// Logger escreve o log no arquivo
func Logger(msg interface{}, tipo string) {
	f, er := os.OpenFile(pathLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if er != nil {
		log.Fatal(er)
	}

	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	_, file, line, _ := runtime.Caller(1)

	/*
		INFO 'e uma mensagem simples(ex: abertura do programa ou inicio do backup). Apresentacao: (x)console (x)log
		TRACE 'e quando ocorre uma acao no sistema. Apresentacao: (x)console ()log
		ERROR 'e quando ocorreu algum erro, mas o programa continua rodando. Apresentacao: (x)console (x)log
		FATAL 'e quando ocorre um erro que 'e necessario parar o programa. Apresentacao: (x)console (x)log
	*/

	if tipo == "info" {
		fmt.Printf("%s - %s\n", time.Now().Format("2006/01/02 15:04:05.000"), msg)
		log.Printf("INFO: %s", msg)
	} else if tipo == "trace" {
		log.Printf("TRACE: %s\n", msg)
	} else if tipo == "error" {
		ContErro++
		fmt.Printf("%s - ERROR: %+v File:%s:%d\n", time.Now().Format("2006/01/02 15:04:05.000"), msg, file, line)
		log.Printf("ERROR: %+v File:%s:%d", msg, file, line)
		arquivoErro(fmt.Sprintf("%s - ERROR: %+v File:%s:%d\n", time.Now().Format("2006/01/02 15:04:05.000"), msg, file, line))
	} else if tipo == "fatal" {
		ContErro++
		fmt.Printf("%s - FATAL: %+v File:%s:%d\n", time.Now().Format("2006/01/02 15:04:05.000"), msg, file, line)
		log.Printf("FATAL: %+v File:%s:%d", msg, file, line)
		arquivoErro(fmt.Sprintf("%s - FATAL: %+v File:%s:%d\n", time.Now().Format("2006/01/02 15:04:05.000"), msg, file, line))
		os.Exit(1) //finaliza a execucao do programa
	}
}

// CheckLogFolder verfica se a pasta com os dados do log esta criada
func CheckLogFolder() {
	ContErro = 0
	if _, err := os.Stat("./log/"); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll("./log/", os.ModePerm); err != nil {
			Logger(err, "fatal")
		}
	}
}

// arquivoErro ira criar um arquivo com os erros identificados durante o processo e o mesmo sera enviado por email ao final da importacao
func arquivoErro(erro string) {
	f, err := os.OpenFile(PathErro, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Logger(err, "error")
		return
	}

	defer f.Close()

	_, err2 := f.WriteString(erro)
	if err2 != nil {
		Logger(err, "error")
		return
	}
}
