# tableconverter

Ferramenta criada para conversão de tabela de dados com variáveis ambientais para o formato Darwin Core.
A conversão se faz necessária devido ao padrão usado pela a maioria dos pesquisadores ser fora do formato exigido pelo DwC-A.

## Comportamento

A ferramenta replica o algoritmo da função *melt* do pacote *reshape2* da linguagem [R](http://www.r-project.org).


Assumindo que uma tabela contem apenas variáveis ambientais o exemplo abaixo em *R* realizada a rotação dos dados para o formato necessário.

```R
# carrega a tabela de dados
# precicsa dos parâmetros completos e o encoding geralmente é 'latin1' ou 'utf8' (preferencia utf8)
dados = read.table('tabela.csv', header = T, sep = ',', dec = '.', encoding = 'utf8')

# carrega a biblioteca que converte a tabela dinamica
library(reshape2)

# cria a coluna eventID nos dados
dados$eventid = 1:nrow(dados)

# passa do formato wide para o long
dadosm = melt(dados, id.vars = 'eventid')
```
