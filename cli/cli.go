package cli

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	ini "github.com/go-ini/ini"
)

// App handles the input from the CLI and encapsulates
// the processing of requests to grant access for an MFA user
// to AWS with temporary STS credentials
type App struct {
	profile   string
	token     string
	region    string
	deviceArn string
	ses       *session.Session
}

// NewApp creates a new app from command line flags
func NewApp(profile, token, region, deviceArn string) *App {
	ses := session.Must(
		session.NewSessionWithOptions(
			session.Options{
				Config:  aws.Config{Region: aws.String(region)},
				Profile: profile,
			},
		),
	)
	return &App{
		profile:   profile,
		token:     token,
		region:    region,
		deviceArn: deviceArn,
		ses:       ses,
	}
}

// SetupUser generates a STS token and updates credentials
func (cli *App) SetupUser() error {
	tokenOutput := cli.generateSTSCreds()
	cli.addProfileWithCreds(tokenOutput)
	return nil
}

func (cli *App) generateSTSCreds() *sts.GetSessionTokenOutput {
	fmt.Printf("Generating STS Token for %s profile\n", cli.profile)
	svc := sts.New(cli.ses)
	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(3600),
		SerialNumber:    aws.String(cli.deviceArn),
		TokenCode:       aws.String(cli.token),
	}

	result, err := svc.GetSessionToken(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeRegionDisabledException:
				fmt.Println(sts.ErrCodeRegionDisabledException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil
	}

	return result
}

func (cli *App) addProfileWithCreds(creds *sts.GetSessionTokenOutput) error {
	awsCreds, err := parseAwsCreds()
	if err != nil {
		return err
	}

	tokenProfile := cli.stsProfile()
	awsCreds.Section(tokenProfile).Key("aws_access_key_id").SetValue(*creds.Credentials.AccessKeyId)
	awsCreds.Section(tokenProfile).Key("aws_secret_access_key").SetValue(*creds.Credentials.SecretAccessKey)
	awsCreds.Section(tokenProfile).Key("aws_session_token").SetValue(*creds.Credentials.SessionToken)
	awsCreds.SaveTo(awsCredsPath())
	fmt.Printf("Adding credentials for %s to %s\n", tokenProfile, awsCredsPath())
	fmt.Printf("Use with --profile=%s or export AWS_DEFAULT_PROFILE=%s\n", tokenProfile, tokenProfile)
	return nil
}

func (cli *App) stsProfile() string {
	return cli.profile + "-sts"
}

func awsCredsPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".aws", "credentials")
}
func parseAwsCreds() (*ini.File, error) {
	creds, err := ini.Load(awsCredsPath())
	return creds, err
}
