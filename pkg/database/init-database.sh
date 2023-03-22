#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER "$POSTGRES_USER" WITH PASSWORD "$POSTGRES_PASSWORD";
    CREATE DATABASE "$POSTGRES_BD" TEMPLATE=template0 LC_CTYPE='pt_BR.UTF-8' LC_COLLATE='pt_BR.UTF-8';
    GRANT ALL PRIVILEGES ON DATABASE "$POSTGRES_BD" TO "$POSTGRES_USER";
    SET TIMEZONE='America/Sao_Paulo';
    CREATE SCHEMA IF NOT EXISTS apac;
    CREATE SCHEMA IF NOT EXISTS geral;
    CREATE TABLE IF NOT EXISTS apac.medicamento (
        competencia_dispensacao date,
        competencia_processamento date,
        codigo_uf_dispensacao numeric(2),
        sigla_uf_dispensacao char(2),
        codigo_municipio_dispensacao numeric(6),
        cnes_dispensacao char(7),
        apac char(13),
        procedimento char(10),
        cid char(4),
        quantidade_apresentada numeric(11),
        quantidade_aprovada numeric(11),
        valor_apresentado numeric(20,2),
        valor_aprovado numeric(20,2),
        cns_paciente char(30),
        idade_paciente numeric(3),
        sexo_paciente char(1),
        raca_cor_paciente char(2),
        etnia_paciente char(4),
        peso_paciente numeric(3),
        altura_paciente numeric(3),
        municipio_residencia_paciente numeric(6),
        nacionalidade_paciente char(3),
        obito_paciente char(1),
        cnes_solicitante char(7),
        data_solicitacao date,
        data_autorizacao date,
        arquivo_origem char(12),
        data_alteracao timestamp default (now() at time zone 'America/Sao_Paulo') not null
    );
    #create index index_competencia_dispensacao ON apac.medicamento (competencia_dispensacao);
    #create index index_codigo_uf_dispensacao ON apac.medicamento (codigo_uf_dispensacao);
    #create index index_procedimento ON apac.medicamento (procedimento);
    #create index index_cid ON apac.medicamento (cid);
    create index index_arquivo_origem ON apac.medicamento (arquivo_origem);
    #create unique index index_unique_apac_competencia_procedimento ON apac.medicamento (apac, competencia_dispensacao, procedimento);
    comment on table apac.medicamento is 'Tabela como os dados de APAC de medicamentos.';
    comment on column apac.medicamento.competencia_dispensacao is 'Referente ao campo AP_CMP. Compentencia arredondada para o primeiro dia do mes em que ocorreu a dispensacao.';
    comment on column apac.medicamento.competencia_processamento is 'Referente ao campo AP_MVM. Compentencia arredondada para o primeiro dia do mes em que ocorreu a importacao do arquivo pela SES para o SIA/SUS.';
    comment on column apac.medicamento.codigo_uf_dispensacao is 'Referente ao campo AP_GESTAO. Dado tratado para conter apenas os dois digitos da UF em que ocorreu a dispensacao.';
    comment on column apac.medicamento.sigla_uf_dispensacao is 'Sigla da UF em que ocorreu a dispensacao.';
    comment on column apac.medicamento.codigo_municipio_dispensacao is 'Referente ao campo AP_UFMUN. Municipio onde ocorreu a dispensacao.';
    comment on column apac.medicamento.cnes_dispensacao is 'Referente ao campo AP_CODUNI. Codigo CNES do estabelecimento dispensador.';
    comment on column apac.medicamento.apac is 'Referente ao campo AP_AUTORIZ. Numero da APAC.';
    comment on column apac.medicamento.procedimento is 'Referente ao campo AP_PRIPAL. Procedimento da SIGTAP dispensado.';
    comment on column apac.medicamento.cid is 'Referente ao campo AP_CIDPRI. CID da APAC.';
    comment on column apac.medicamento.quantidade_apresentada is 'Referente ao campo PA_QTDPRO do arquivo AP. Quantidade apresentada pela SES para o procedimento.';
    comment on column apac.medicamento.valor_apresentado is 'Referente ao campo PA_VALPRO. Valor financeiro apresentado pela SES para o procedimento.';
    comment on column apac.medicamento.quantidade_aprovada is 'Referente ao campo PA_QTDAPR do arquivo AP. Quantidade aprovada pelo SIA/SUS para o procedimento.';
    comment on column apac.medicamento.valor_aprovado is 'Referente ao campo AP_VL_AP. Valor financeiro aprovado para o procedimento.';
    comment on column apac.medicamento.cns_paciente is 'Referente ao campo AP_CNSPCN. CNS do paciente codificado em hexadecimal. Cada caractere de criptografia do Datasus equivale a dois caracteres no banco.';
    comment on column apac.medicamento.idade_paciente is 'Referente ao campo AP_NUIDADE. Idade do paciente no momento da dispensacao.';
    comment on column apac.medicamento.sexo_paciente is 'Referente ao campo AP_SEXO. Sexo do paciente.';
    comment on column apac.medicamento.raca_cor_paciente is 'Referente ao campo AP_RACACOR. Raca/cor do paciente. Opcoes: 01-Branca, 02-Preta, 03-Parda, 04-Amarela, 05-Indígena, 99-Sem informação.';
    comment on column apac.medicamento.etnia_paciente is 'Referente ao campo AP_ETNIA. Etnia do paciente.';
    comment on column apac.medicamento.peso_paciente is 'Referente ao campo AM_PESO. Peso do paciente informado no LME.';
    comment on column apac.medicamento.altura_paciente is 'Referente ao campo AM_ALTURA. Altura do paciente em cm informado no LME.';
    comment on column apac.medicamento.municipio_residencia_paciente is 'Referente ao campo AP_MUNPCN. Municipio de residencia do paciente.';
    comment on column apac.medicamento.nacionalidade_paciente is 'Referente ao campo AP_UFNACIO. Nacionalidade do paciente.';
    comment on column apac.medicamento.obito_paciente is 'Referente ao campo AP_OBITO. Dado tratado para boolean. Indica se paciente possui obito informado na APAC.';
    comment on column apac.medicamento.cnes_solicitante is 'Referente ao campo AP_UNISOL. CNES do profissional solicitante do LME.';
    comment on column apac.medicamento.data_solicitacao is 'Referente ao campo AP_DTSOLIC. Data de solicitacao do LME.';
    comment on column apac.medicamento.data_autorizacao is 'Referente ao campo AP_DTAUT. Data de autoriazacao do LME.';
    comment on column apac.medicamento.arquivo_origem is 'Arquivo do FTP do Datasus em que o registro foi extraido.';
    CREATE TABLE IF NOT EXISTS apac.temp_medicamento (
        competencia_dispensacao date,
        competencia_processamento date,
        codigo_uf_dispensacao numeric(2),
        sigla_uf_dispensacao char(2),
        codigo_municipio_dispensacao numeric(6),
        cnes_dispensacao char(7),
        apac char(13),
        procedimento char(10),
        cid char(4),
        quantidade_apresentada numeric(11),
        quantidade_aprovada numeric(11),
        valor_apresentado numeric(20,2),
        valor_aprovado numeric(20,2),
        cns_paciente char(30),
        idade_paciente numeric(3),
        sexo_paciente char(1),
        raca_cor_paciente char(2),
        etnia_paciente char(4),
        peso_paciente numeric(3),
        altura_paciente numeric(3),
        municipio_residencia_paciente numeric(6),
        nacionalidade_paciente char(3),
        obito_paciente char(1),
        cnes_solicitante char(7),
        data_solicitacao date,
        data_autorizacao date,
        arquivo_origem char(12),
        data_alteracao timestamp default (now() at time zone 'America/Sao_Paulo') not null
    );
    create index index_apac ON apac.temp_medicamento (apac);
    comment on table apac.temp_medicamento is 'Tabela temporaria apenas como os dados de APAC do arquivo AM do Datasus.';
    comment on column apac.temp_medicamento.competencia_dispensacao is 'Referente ao campo AP_CMP. Compentencia arredondada para o primeiro dia do mes em que ocorreu a dispensacao.';
    comment on column apac.temp_medicamento.competencia_processamento is 'Referente ao campo AP_MVM. Compentencia arredondada para o primeiro dia do mes em que ocorreu a importacao do arquivo pela SES para o SIA/SUS.';
    comment on column apac.temp_medicamento.codigo_uf_dispensacao is 'Referente ao campo AP_GESTAO. Dado tratado para conter apenas os dois digitos da UF em que ocorreu a dispensacao.';
    comment on column apac.temp_medicamento.sigla_uf_dispensacao is 'Sigla da UF em que ocorreu a dispensacao.';
    comment on column apac.temp_medicamento.codigo_municipio_dispensacao is 'Referente ao campo AP_UFMUN. Municipio onde ocorreu a dispensacao.';
    comment on column apac.temp_medicamento.cnes_dispensacao is 'Referente ao campo AP_CODUNI. Codigo CNES do estabelecimento dispensador.';
    comment on column apac.temp_medicamento.apac is 'Referente ao campo AP_AUTORIZ. Numero da APAC.';
    comment on column apac.temp_medicamento.procedimento is 'Referente ao campo AP_PRIPAL. Procedimento da SIGTAP dispensado.';
    comment on column apac.temp_medicamento.cid is 'Referente ao campo AP_CIDPRI. CID da APAC.';
    comment on column apac.temp_medicamento.quantidade_apresentada is 'Referente ao campo PA_QTDPRO do arquivo AP. Quantidade apresentada pela SES para o procedimento.';
    comment on column apac.temp_medicamento.valor_apresentado is 'Referente ao campo PA_VALPRO. Valor financeiro apresentado pela SES para o procedimento.';
    comment on column apac.temp_medicamento.quantidade_aprovada is 'Referente ao campo PA_QTDAPR do arquivo AP. Quantidade aprovada pelo SIA/SUS para o procedimento.';
    comment on column apac.temp_medicamento.valor_aprovado is 'Referente ao campo AP_VL_AP. Valor financeiro aprovado para o procedimento.';
    comment on column apac.temp_medicamento.cns_paciente is 'Referente ao campo AP_CNSPCN. CNS do paciente codificado em hexadecimal. Cada caractere de criptografia do Datasus equivale a dois caracteres no banco.';
    comment on column apac.temp_medicamento.idade_paciente is 'Referente ao campo AP_NUIDADE. Idade do paciente no momento da dispensacao.';
    comment on column apac.temp_medicamento.sexo_paciente is 'Referente ao campo AP_SEXO. Sexo do paciente.';
    comment on column apac.temp_medicamento.raca_cor_paciente is 'Referente ao campo AP_RACACOR. Raca/cor do paciente. Opcoes: 01-Branca, 02-Preta, 03-Parda, 04-Amarela, 05-Indígena, 99-Sem informação.';
    comment on column apac.temp_medicamento.etnia_paciente is 'Referente ao campo AP_ETNIA. Etnia do paciente.';
    comment on column apac.temp_medicamento.peso_paciente is 'Referente ao campo AM_PESO. Peso do paciente informado no LME.';
    comment on column apac.temp_medicamento.altura_paciente is 'Referente ao campo AM_ALTURA. Altura do paciente em cm informado no LME.';
    comment on column apac.temp_medicamento.municipio_residencia_paciente is 'Referente ao campo AP_MUNPCN. Municipio de residencia do paciente.';
    comment on column apac.temp_medicamento.nacionalidade_paciente is 'Referente ao campo AP_UFNACIO. Nacionalidade do paciente.';
    comment on column apac.temp_medicamento.obito_paciente is 'Referente ao campo AP_OBITO. Dado tratado para boolean. Indica se paciente possui obito informado na APAC.';
    comment on column apac.temp_medicamento.cnes_solicitante is 'Referente ao campo AP_UNISOL. CNES do profissional solicitante do LME.';
    comment on column apac.temp_medicamento.data_solicitacao is 'Referente ao campo AP_DTSOLIC. Data de solicitacao do LME.';
    comment on column apac.temp_medicamento.data_autorizacao is 'Referente ao campo AP_DTAUT. Data de autoriazacao do LME.';
    comment on column apac.temp_medicamento.arquivo_origem is 'Arquivo do FTP do Datasus em que o registro foi extraido.';
    CREATE TABLE IF NOT EXISTS apac.auditoria (
        arquivo varchar(64) not null primary key,
        registros_dbf numeric not null,
        registros_lidos numeric,
        registros_apac_medicamento numeric,
        registros_banco numeric,
        inicio_processo timestamp not null,
        fim_processo timestamp,
        data_alteracao timestamp default (now() at time zone 'America/Sao_Paulo') not null
    );
    comment on table apac.auditoria is 'Tabela com as informacoes do processo de carga dos arquivos do Datasus.';
    comment on column apac.auditoria.arquivo is 'Nome do arquivo DBC importado do FTP do Datasus.';
    comment on column apac.auditoria.registros_dbf is 'Quantidade total de registros no arquivo DBF do tipo PA. Inclui registros nao CEAF';
    comment on column apac.auditoria.registros_lidos is 'Quantidade total de registros que o programa leu. Deve ser o mesmo valor do campo registros_dbf.';
    comment on column apac.auditoria.registros_apac_medicamento is 'Quantidade total de registros do tipo APAC de medicamento lidos pelo programa.';
    comment on column apac.auditoria.registros_banco is 'Quantidade de registros persistidos no banco de dados com os dados PA. Deve ser o mesmo valor do campo registros_apac_medicamento';
    comment on column apac.auditoria.inicio_processo is 'Data que o iniciou o processo de obter o arquivo DBC.';
    comment on column apac.auditoria.fim_processo is 'Data que o processo de persistencia foi finalizado.';
    comment on column apac.auditoria.data_alteracao is 'Data da ultima atualizacao no registro.';
    
EOSQL