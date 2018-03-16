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
	duration := flag.Int64("duration", cli.DefaultDuration, "Time in seconds the token is valid for (max 24 hours e.g. 86400)")
	flag.Parse()

	inputArgs := &cli.InputArgs{
		Profile:   *profile,
		Token:     *token,
		Region:    *region,
		DeviceArn: *deviceArn,
		Duration:  *duration,
	}
	inputArgs.ValidateFields()

	app := cli.NewApp(inputArgs)
	app.SetupUser()
}
