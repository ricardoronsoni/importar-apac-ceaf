package database

import (
	"dados-apac/pkg/utils"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DbConnection realiza a conexao com o banco de dados
func DbConnection() *sql.DB {
	//a conexao com o banco de dados no Docker será diferente se a aplicacao Go estiver rodando no docker ou localmente
	connectionStringGoDocker := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", utils.DbHost, utils.DbPort, utils.DbUser, utils.DbPassword, utils.DbName)
	connectionStringGoLocal := fmt.Sprintf("host=127.0.0.1 port=%s user=%s password=%s dbname=%s sslmode=disable", utils.DbPort, utils.DbUser, utils.DbPassword, utils.DbName)

	db := conectarDb(connectionStringGoDocker)

	//defer db.Close()

	err := db.Ping()
	if err != nil {
		db = conectarDb(connectionStringGoLocal)
		err := db.Ping()
		if err != nil {
			utils.Logger(err, "fatal")
		}
	}

	return db
}

func conectarDb(stringocnexao string) *sql.DB {
	db, err := sql.Open("postgres", stringocnexao)
	if err != nil {
		utils.Logger(err, "fatal")
	}

	return db
}
