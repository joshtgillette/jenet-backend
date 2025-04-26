package main

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	path := strings.TrimPrefix(request.RawPath, "/release")

	switch path {
	case "/tagline":
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Body:       "Generated ui for <i>you</i>, coming soon",
		}, nil
	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Body:       "Page not found",
		}, nil
	}
}

func main() {
	lambda.Start(handler)
}
