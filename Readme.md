This is beginnerd, a simple log collector to S3 using Kinesis Firehose.

Configure your credentials like
```
export AWS_REGION=us-west-2
export AWS_ACCESS_KEY_ID=YOUR_AKID
export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
```

Execute
```
# Make sure your $GOPATH is set
git clone https://github.com/pfeodrippe/beginnerd.git $GOPATH/src/github.com/pfeodrippe/beginnerd
# or (not tested) go get github.com/pfeodrippe/beginnerd
cd $GOPATH/src/github.com/pfeodrippe/beginnerd
go get
go build
go install
```
You'll have a new command: ```beginnerd``` or ```./beginnerd``` =D, see more at ```beginnerd --help```.

See the ```terraform/Readme.md``` and ```log-synth/Readme.md``` files for more instructions and execute below to send the logs to Kinesis Firehose.

After you follow and execute the log-synth and terraform instructions, execute 
```
beginnerd agglogs -f log-synth/output/synth-0000..json
```
