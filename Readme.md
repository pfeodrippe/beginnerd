This is beginnerd, a simple log collector to S3 using Kinesis Firehose.

Configure your credentials like
```
export AWS_REGION=us-west-2
export AWS_ACCESS_KEY_ID=YOUR_AKID
export AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
```

Execute
```
git clone https://github.com/pfeodrippe/beginnerd.git
# or (not tested) go get github.com/pfeodrippe/beginnerd
cd beginnerd
go get
go install
```
You'll have a new command: ```beginnerd``` =D, see ```beginnerd --help```.

See the ```terraform/Readme.md``` and ```log-synth/Readme.md``` files for more instructions.

After you follow and execute the log-synth and terraform instructions, execute 
```
beginnerd agglogs -f log-synth/output/synth-0000..json
```