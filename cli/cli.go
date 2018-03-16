package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	ini "github.com/go-ini/ini"
)

// DefaultDuration is the default duration for a token to be valid (1hr) in seconds
const DefaultDuration int64 = 3600

// App handles the input from the CLI and encapsulates
// the processing of requests to grant access for an MFA user
// to AWS with temporary STS credentials
type App struct {
	profile   string
	token     string
	region    string
	deviceArn string
	duration  int64
	ses       *session.Session
}

// InputArgs takes in flag inputs from the command line to be passed into
// a NewApp to create the App with user specificed inputs
type InputArgs struct {
	Profile   string
	Token     string
	Region    string
	DeviceArn string
	Duration  int64
}

// NewApp creates a new app from command line flags
func NewApp(input *InputArgs) *App {
	ses := session.Must(
		session.NewSessionWithOptions(
			session.Options{
				Config:  aws.Config{Region: aws.String(input.Region)},
				Profile: input.Profile,
			},
		),
	)
	return &App{
		profile:   input.Profile,
		token:     input.Token,
		region:    input.Region,
		deviceArn: input.DeviceArn,
		duration:  input.Duration,
		ses:       ses,
	}
}

// SetupUser generates a STS token and updates credentials
func (cli *App) SetupUser() error {
	if tokenOutput := cli.generateSTSCreds(); tokenOutput != nil {
		return cli.addProfileWithCreds(tokenOutput)
	}
	return errors.New("Unable to generate a token with provided credentials")
}

func (cli *App) generateSTSCreds() *sts.GetSessionTokenOutput {
	fmt.Printf("Generating STS Token for %s profile\n", cli.profile)
	svc := sts.New(cli.ses)
	input := &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(cli.duration),
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
	fmt.Printf("Use with the awscli by passing --profile=%s \n\n", tokenProfile)
	fmt.Printf("or set up your environment with \n\nexport AWS_DEFAULT_PROFILE=%s\nexport AWS_PROFILE=%s\n", tokenProfile, tokenProfile)
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

// ValidateFields will check that required fields are passed in from input
// and ensure the duration is in the valid range
func (inputs *InputArgs) ValidateFields() {
	if inputs.Token == "" || len(inputs.Token) < 6 {
		fmt.Println("Token required and must be at least 6 digits\n")
		flag.Usage()
		os.Exit(1)
	}
	if inputs.DeviceArn == "" {
		fmt.Println("MFA Device ARN required. Please go to your IAM user and copy.\n")
		flag.Usage()
		os.Exit(1)
	}
	if inputs.Duration < 900 || inputs.Duration > 86400 {
		inputs.Duration = DefaultDuration
		fmt.Printf("Invalid duration %d; setting to default of 1hr\n", inputs.Duration)
	}
}
