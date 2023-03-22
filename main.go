package main

import (
	"dados-apac/pkg/services"
	"dados-apac/pkg/utils"
	"fmt"
	"syscall"
	"time"
)

func main() {
	utils.CarregarEnv()
	utils.CheckLogFolder()
	services.DeletarArquivosTemporarios()
	fmt.Printf("%s - Monitor APAC iniciado. Atual PID: %d. \n", time.Now().Format("2006/01/02 15:04:05.000"), syscall.Getpid())
	services.Cron()
}
