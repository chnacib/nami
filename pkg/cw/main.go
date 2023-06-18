package cw

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func CpuAverage(cluster string, service string) float64 {
	var utilization float64
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := cloudwatch.New(sess)

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/ECS"),
		MetricName: aws.String("CPUUtilization"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("ClusterName"),
				Value: aws.String(cluster),
			},
			{
				Name:  aws.String("ServiceName"),
				Value: aws.String(service),
			},
		},
		StartTime: aws.Time(time.Now().Add(-1 * 24 * time.Hour)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(300),
		Statistics: []*string{
			aws.String(cloudwatch.StatisticAverage),
		},
	}

	result, err := svc.GetMetricStatistics(input)
	if err != nil {
		utilization = 0.0
		fmt.Println(err)
	}

	if len(result.Datapoints) > 0 {
		utilization = aws.Float64Value(result.Datapoints[0].Average)
	} else {
		utilization = 0.0
	}

	return utilization
}

func MemoryAverage(cluster string, service string) float64 {
	var utilization float64
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := cloudwatch.New(sess)

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/ECS"),
		MetricName: aws.String("MemoryUtilization"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("ClusterName"),
				Value: aws.String(cluster),
			},
			{
				Name:  aws.String("ServiceName"),
				Value: aws.String(service),
			},
		},
		StartTime: aws.Time(time.Now().Add(-1 * 24 * time.Hour)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(300),
		Statistics: []*string{
			aws.String(cloudwatch.StatisticAverage),
		},
	}

	result, err := svc.GetMetricStatistics(input)
	if err != nil {
		utilization = 0.0
		fmt.Println(err)
	}

	if len(result.Datapoints) > 0 {
		utilization = aws.Float64Value(result.Datapoints[0].Average)
	} else {
		utilization = 0.0
	}

	return utilization
}
