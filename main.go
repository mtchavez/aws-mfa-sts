package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mtchavez/aws-mfa-sts/cli"
)

func main() {
	profile := flag.String("profile", "default", "Profile name")
	token := flag.String("token", "", "MFA token")
	region := flag.String("region", "us-east-1", "AWS Region")
	deviceArn := flag.String("device-arn", "", "MFA Serial Number")
	flag.Parse()
	if *token == "" {
		fmt.Println("Token required")
		flag.Usage()
		os.Exit(1)
	}
	if *deviceArn == "" {
		fmt.Println("MFA Device ARN required. Please go to your IAM user and copy.")
		flag.Usage()
		os.Exit(1)
	}
	app := cli.NewApp(*profile, *token, *region, *deviceArn)
	app.SetupUser()
}
