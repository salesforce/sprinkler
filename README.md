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

Start control service on a new tab, to later post a Sprinkler json workflow
```
go run . service control --config .sprinkler.yaml 
```

Put to `http://localhost:8080/v1/workflow` a Sprinkler workflow, example payload from [ExampleWorkflow.scala](https://github.com/salesforce/spade/blob/main/spade-examples/src/main/scala/com/salesforce/mce/spade/examples/ExampleWorkflow.scala)
```
{
    "name": "test-workflow",
    "artifact": "s3://sprinkler-salesforce-bucket/jars/test/spade-example.jar",
    "command":  "[\"java\", \"-cp\", \"spade-example.jar\", \"com.salesforce.mce.spade.examples.ExampleWorkflow\", \"generate\", \"--compact\"]",
    "every": "60.minute",
    "nextRuntime": "2024-08-22T21:00:00Z",
    "backfill": false,
    "owner": "arn:aws:sns:us-east-2:444455556666:MyTopic",
    "isActive": true
}
```

Start fake orchard service
```
go run . service orchard
```
This will start fake orchard service listening on port `8082` based on `.sprinkler.yaml` scheduler orchardAddress.

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
