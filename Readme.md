This is beginnerd, a simple log collector to S3 using Kinesis Firehose.

Configure your credentials like
```
export AWS_REGION=us-west-2
export AWS_ACCESS_KEY_ID=YOUR_AKID
export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
```

Execute
```
go get
go install
```
You'll have a new command: beginnerd.


See the ```terraform/Readme.md``` and ```log-synth/Readme.md``` files for more instructions.