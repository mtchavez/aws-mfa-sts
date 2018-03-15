package main

import (
	"flag"

	"github.com/mtchavez/aws-mfa-sts/cli"
)

func main() {
	profile := flag.String("profile", "default", "Profile name")
	token := flag.String("token", "", "MFA token")
	region := flag.String("region", "us-east-1", "AWS Region")
	deviceArn := flag.String("device-arn", "", "MFA Serial Number")
	flag.Parse()
	app := cli.NewApp(*profile, *token, *region, *deviceArn)
	app.SetupUser()
}
