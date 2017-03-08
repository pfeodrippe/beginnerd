Create a secret.tfvars file at this folder and and put your credentials.
```
vi secret.tfvars
```
```
access_key = "XXXXXXXXXXXXXXXXXXXXXXXXX"
secret_key = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
region     = "us-east-1" # or any other region you want
```

Configure the main.tf file to change names at your will. If you find some error, try changing thes S3 bucket name to a unique name
