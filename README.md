# sprinkler

A scheduler service for [orchard](https://github.com/salesforce/orchard)

## Get Started
Build
```
go build
```
Set db password
```
export POSTGRES_PASSWORD=sprinkler
```
Start the docker container
```
docker compose up -d
```
Start web service
```
docker compose exec ws bash
```
Initilize the database with some sample data
```
go run . database initialize --with-sample --config .sprinkler.yaml
```
Log into database to check whether the tables are created and populated with some sample data.
```
docker compose exec db psql -h db -U postgres
```
Start fake orchard service
```
go run . service orchard
```
This will start fake orchard service listening on port 8081.

Open another tab to run the scheduler.
```
go run . service scheduler --config .sprinkler.yaml
```
The scheduler should start scheduling the sample workflows.


## Contribute

```
# Add licence headers for new source code files
~/go/bin/addlicense -f LICENSE_HEADER.txt ./**/*.go
```
