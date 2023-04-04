# Dados APAC CEAF
Sistema para coletar os dados de APAC dos medicamentos do Componente Especializado da Assistência Farmacêutica (CEAF) disseminados pelo Datasus. Por meio de rotina automatizada os dados são coletados, tratados, validados e persistidos em banco de dados.

# Principais funcionalidades
- Importa os dados de APAC dos medicamentos do CEAF.  
- Salva os dados em banco de dados.  
- Multiplataforma (Windows, MacOS e Linux).  
- Ambiente conteinerizado funcional apenas com a instalação do software Docker.  
- Mantém o banco de dados sempre atualizado com base em uma rotina automática e configurável de atualização dos registros.  
- Envio de emails sempre que a rotina de atualização for iniciada e finalizada com o resultado da rotina.  
- Vários fluxos para manter a integridade do banco de dados e evitar a ocorrência de falta de registros ou duplicidade.  

# Tecnologias utilizadas
- Linguagem de programação: Golang 1.20.2.
- Banco de dados: Postgres 15.2.
- Container: Docker.

# Iniciar a aplicação
## Instalar o software Docker
O sistema e o seu banco de dados rodam em um container, e por isso necessitam apenas da instalação do software [Docker](https://docs.docker.com/get-docker/) para o seu pleno funcionamento.
## Iniciar o container
Acessar a pasta raiz do projeto que contêm o arquivo `dockerfile-compose.yml` e executar o seguinte comando no terminal: 
```
docker-compose up
```
A partir disso será possível acompanhar no terminal todas as ações que estão sendo realizadas pelo programa. 
Se ocorrer algum problema de permissão de diretório e o banco de dados não for criado no Docker, será necessário acessar o container do postgres da seguinte forma.  
```
docker exec -it postgres-analise-dados psql -U <usuario_do_banco_especificado_no_.env>
```
Após isso é só inserir manualmente as instruções SQL disponíveis em `./pkg/database/init-database.sh`.  
Caso seja necessário alterar algum parâmetro de configuração do sistema, o container deverá ser stopado e o código abaixo deverá ser executado para que as novas definições sejam aplicadas.
```
docker-compose up --build
```

# Configurações do sistema
O programa possui algumas configurações que os usuários devem realizar, conforme abaixo.
## Variáveis de ambiente
No arquivo `.env-example` situado na pasta raiz do programa deverá ser informado o usuário e senha do email que será utilizado para enviar os e-mails durante a execução do programa. Caso queira, também poderá ser alterado os parâmetros do banco de dados (não obrigatório). Posteriormente, o arquivo `.env-example` deverá ser renomeado para `.env`.
## Parametrização programa
No arquivo `./pkg/utils/config.go` é possível parametrizar o sistema, conforme abaixo:
- Periodicidade para buscar atualizações no diretório do Datasus.
- Data inicial e final para buscar os dados.
- Configurações para o envio de email.

# Informações
## Log
Na pasta `./log` é disponibilizado o log com as ações do sistema durante a sua execução. Para cada dia de execução é criado um arquivo. Também é disponibilizado um arquivo no mesmo diretório do log para compilar os erros durante a execução do programa. Arquivo também é segmentado por dia.

## SQL
Todas as instruções DDL e DML necessárias para o sistema são carregados durante a criação do container. Os mesmos estão disponíveis em `./pkg/database/init-database.sh`.  
Schema com os dados das APAC: `apac`.  
Schema com tabela de UF e município: `geral`. 

## DBC para DBF
Foi utilizada a biblioteca https://github.com/eaglebh/blast-dbf.git para realizar a transformação dos arquivos DBC para DBC. O fluxo para utilizar a biblioteca manualmente é:
```
git clone https://github.com/eaglebh/blast-dbf.git
cd blast-dbf
make
./blast-dbf input.dbc output.dbf
```
A biblioteca está disponível para o sistema na pasta `./dbc/dbcParaDbf` do projeto.

# Requisitos não funcionais
## RNF001 - Facilidade de instalação
O programa e o banco de dados deverão estar conteinerizados, possibilitando a utilização do mesmo apenas com a instalação do software Docker.
## RNF002 - Performance satisfatória
A performance do programa deve ser satisfatória para que os milhões de registros sejam importados, tratados, gravados e validados com rapidez.
## RNF003 - Integridade do banco de dados 
A integridade do banco de dados é requisito indispensável ao programa, para que não faltem registros, ocorra duplicidade dos dados ou persistência de informações incompletas ou com erros.
## RF004 - Multiplataforma
O programa deve ser funcional nos sistemas operacionais Windows, Linux e MacOS.


# Requisitos funcionais
## RN001 - Agendamento
O sistema deverá permitir que o agendamento seja parametrizável pelo usuário (ex: todo dia, dias alternados, semanalmente, horário específico, ...), a partir do arquivo de configurações do programa.
## RN002 - Sobreposição de execuções
O programa não poderá iniciar um novo agendamento caso a execução do agendamento anterior não tenha sido finalizado.
## RN003 - Range de histórico
O usuário deverá informar no arquivo de configuração a data inicial e final que a rotina irá buscar os dados de APAC. A partir desses dados serão importados os arquivos DBC do Datasus que possuem em seu nome a competência dentro do range (ex: AMAC2101.dbc).
## RN004 - Visualização rotina
A rotina deverá mostrar no console do sistema todos os passos que estão sendo realizados, em conjunto com o horário de cada passo.
## RN005 - Log
O programa deverá gravar todos os eventos e erros em arquivo de log. Será criado um arquivo de log para cada dia de execução, sendo que a data fará parte do nome do arquivo para organização.
## RN006 - Arquivo de erros
Para facilitar a visualização de possíveis erros no log, será criado um arquivo no mesmo diretório do log para compilar os erros durante a execução do programa. Arquivo será segmentado por dia.
## RN007 - Notificação por email
O sistema deverá notificar os usuários por email nas seguintes situações:  
    - Agendamento não localizou novos arquivos.  
    - Agendamento localizou novos arquivos. Nesse caso, deverão ser listados os arquivos localizados para importação e informado que o processo de persistência dos arquivos foi iniciado.  
    - Término do processo de importação dos arquivos localizados anteriormente, informando a quantidade de erros e de arquivos persistidos com sucesso. 
## RN008 - Parametrização envio email
O usuário poderá informar no arquivo de configuração do sistema os destinatários que irão receber os emails do sistema. Também deverá ser configurado um usuário e senha para o envio dos emails no arquivo `.env`.
## RN009 - Escopo 
Serão persistidos apenas os dados dos medicamentos do Componente Especializado da Assistência Farmacêutica (CEAF), para as 27 unidades da federação. 
## RN010 - Auditoria
Sistema terá uma tabela de auditoria para gravar os dados durante o processo de importação. O preenchimento do campo `fim_processo` determina que o arquivo foi devidamente salvo no banco de dados.
## RN011 - Validação inicial auditoria
Ao iniciar a importação de um arquivo DBC o programa deverá verificar se existe algum arquivo DBC na tabela de auditoria sem o preenchimento do campo `fim_processo`. Em havendo, os dados desse arquivo DBC nas tabelas de auditoria e de medicamento deverão ser excluídos para que possa ocorrer a importação novamente desse arquivo DBC.
## RN012 - Validação inicial duplicidade
Antes de iniciar o processo de importação de arquivo DBC o sistema deverá consultar a tabela de medicamento e verificar se o mesmo possui algum dado persistido. Se houver, para evitar duplicidade todos os dados da tabela de medicamento e de auditoria deverão ser excluídos para o respectivo arquivo DBC. Após, o arquivo poderá ser importado.
## RN013 - Exclusão arquivos temporários
A medida que o processo de importação for ocorrendo os arquivos temporários nas extensões .DBC, .DBF e .CSV deverão ser excluídos. Ao final e inicio da execução do programa o sistema também deverá rodar rotina para apagar todos os arquivos temporários.
## RN014 - Ordem persistência
Primeiramente, deverão ser gravados os registros dos arquivos do tipo 'PA', conforme a RN008. Posteriormente, esses registros deverão ser atualizados com os dados do arquivo do tipo 'AM'.
## RN015 - Complementação registro ausente
Caso o arquivo 'AM' da respectiva competência não possua os dados de algum registro do tipo 'PA', o programa deverá consultar no histórico do banco de dados os dados da respectiva APAC para complementar as informações.
## RN016 - Verificação de erros
Se o programa identificar qualquer tipo de erro durante a sua execução o processo de importação do respectivo arquivo deverá ser suspenso, bem como as suas informações do banco de dados serão excluídas. Contudo, a rotina irá continuar importando os demais arquivos pendentes para as demais competências e/ou UF. O arquivo que originou o problema será importado no próximo agendamento. 
## RN017 - Erros fatais
Erros que ocorrerem durante alguma validação da consistência do banco de dados serão classificados como fatais, provocando o término de execução por completo do programa, inclusive os próximos agendamento. Esses erros estarão disponíveis no arquivo de log, conforme a RN005.
## RN018 - Validar tamanho arquivo
Ao realizar o download de um arquivo o sistema deverá comparar o tamanho no disco local com o tamanho do servidor FTP do Datasus. Em havendo discrepância, o processo de importação do arquivo será suspenso.
## RN019 - Validar quantidade de registros
Ao realizar o download de um arquivo do tipo 'PA', deverá ser contado a quantidade de registros contidos nele. Isto deve ocorrer antes de iniciar qualquer tipo de manipulação no arquivo e devem ser contados todos os registros, inclusive os não CEAF. Posteriormente, durante a manipulação do arquivo deverá ser realizada outra contagem para somar os registros lidos pelo programa. Caso as contagens não batam o processo de importação do arquivo será suspenso.
## RN020 - Validar quantidade de registros CEAF
Durante o processo de importação o sistema deverá contar a quantidade de registros do CEAF que foram identificados no arquivo 'PA'. Após a finalização da importação, o programa deverá contar a quantidade de registros do CEAF inseridos no banco de dados para o respectivo arquivo 'PA'. Caso as contagens não batam o processo de importação do arquivo será suspenso e todos os registros no banco de dados para o respectivo arquivo serão excluídos.
## RN021 - Data atualização
Todo registro no banco deverá possuir preenchido a data da sua última atualização.
## RN022 - Dados inconsistentes
Os dados que forem recebidos fora do padrão ou em branco deverão ser gravados no banco de dados como Null. Nesse caso o registro deverá ser persistido no banco, exceto se o problema ocorrer em um campos listados abaixo, onde todos os registros do arquivo DBC serão rejeitados:
    - Competência da dispensação;
    - Competência do processamento;
    - Código UF dispensação;
    - Quantidade aprovada;
    - Valor aprovado.

# Benchmark
Abaixo é disponibilizado o tempo de execução do programa para realizar as importações. Teste realizado com Mackbook Air M1, 16Gb ram, 256Gb SSD.
- Tempo para importar o ano de 2021 completo: 13h13m16s726.
- Tempo para importar individualmente as competências de 2021: Avg: 01h06m06s393, Min: 00h52m29s185, Max: 01h26m09s654.

# Contato
## ricardoronsoni@gmail.com