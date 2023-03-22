package services

import (
	"dados-apac/pkg/utils"
	"os"
	"os/signal"

	"github.com/robfig/cron/v3"
)

// Cron agenda as rotinas de pesquisa no Datasus
func Cron() {
	cronInicio()
	listen()
}

// cronInicio inicia o processo de agendamento
func cronInicio() {
	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger), //ira evitar a sobreposicao de execucoes
	))

	id, _ := c.AddFunc(utils.AgendamentoDatasus, IniciarMonitor)
	c.Entry(id).Job.Run()

	//para iniciar outro agendamento necessario
	//c.AddFunc(utils.ScheduleStillAlive, emailStillAlive)

	go c.Start()
}

// listen eh uma funcao da biblioteca cron necessaria para o agendamento
func listen() {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
}
