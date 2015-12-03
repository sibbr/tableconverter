# tableconverter

Tool created to convert data table with environment variables to Darwin Core standard.
The conversion is necessary because the common layout used by researchers (wide format) does not match DwC layout.

## Behavior

This tool is just replicating the *melt* algorithm from package [reshape2](https://cran.r-project.org/web/packages/reshape2/index.html) from R language.

The R script below is how to achieve two columns required by the measurement and facts extension ([measurementType](http://rs.tdwg.org/dwc/terms/#measurementType), [measurementValue](http://rs.tdwg.org/dwc/terms/#measurementValue)).

```R
# load data
# take extra care with encoding, most common for us are 'latin1' ou 'utf8' (please use utf8 if possible)
data = read.table('table.csv', header = T, sep = ',', dec = '.', encoding = 'utf8')

# load the library needed for melt function
library(reshape2)

# create a eventid column to maintain the original relationship of lines
data$eventid = 1:nrow(data)

# conversion from wide to long format
dataMelted = melt(data, id.vars = 'eventid')
```

## Why create tableconverter if R can do the job?

Our experience shows users does not want to learn or install a new tool. R is pretty popular among ecologists but we needed to access a wide audience.
Teaching how to use R or pivot tables aren't our focus and will take precious time we needed.

At time of writing this README our tool has only two steps, select a table and which columns to not "melt".

## How to compile

We provide binaries on SiBBr courses but if you want to compile follow the steps described below.

- Install and configure an environment for Go. [Follow this guide.](https://golang.org/doc/install)
- If you did everything right type:
```
$ go get -d github.com/sibbr/tableconverter
$ cd $GOPATH/src/github.com/sibbr/tableconverter
$ go build
```
- Now you should have a binary called tableconverter[.exe]

## How to use

Execute tableconverter, no output will be show if everything went ok.

- Open your browser at http://localhost:8080
- Select the table you want to convert (need to be in csv and UTF-8)
- Select the field separator
- Press Send
- Select which columns you want to remain fixed (usually terms of DwC)
- Press Convert and download the new file

## Credits

[David Dias](https://github.com/dvdscripter): main developer

[Eduardo Rudas](https://github.com/erudas-SiBBr): algorithm suggestions

[Jurandir Júnior](https://github.com/jurajunior): main designer
