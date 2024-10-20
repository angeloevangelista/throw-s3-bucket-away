package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("sdk resolveu piar agora, %v", err)
	}

	svc := s3.NewFromConfig(cfg)

	bucketName := "some-greate-bucket"

	deleteObjectVersions := func() {
		listObjectVersionsInput := &s3.ListObjectVersionsInput{
			Bucket: aws.String(bucketName),
		}

		paginator := s3.NewListObjectVersionsPaginator(svc, listObjectVersionsInput)

		counter := 0

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(context.TODO())
			if err != nil {
				log.Fatalf("a proxima pagina sumiu: %v", err)
			}

			waitGroup := sync.WaitGroup{}
			totalItems := len(page.Versions) + len(page.DeleteMarkers)
			waitGroup.Add(totalItems)

			for _, version := range page.Versions {
				go func(version types.ObjectVersion) {
					deleteObjectInput := &s3.DeleteObjectInput{
						Bucket:    aws.String(bucketName),
						Key:       version.Key,
						VersionId: version.VersionId,
					}
					_, err := svc.DeleteObject(context.TODO(), deleteObjectInput)
					if err != nil {
						log.Printf("nao deu pra excluir o '%s', vou fingir q nao vi: %v", *version.Key, err)
					} else {
						fmt.Printf("%s (%s) foi cinzado\n", *version.Key, *version.VersionId)
					}

					waitGroup.Done()
					counter++
					fmt.Printf("matamo '%d' coisas até agora\n", counter)
				}(version)
			}

			for _, marker := range page.DeleteMarkers {
				go func(marker types.DeleteMarkerEntry) {
					deleteObjectInput := &s3.DeleteObjectInput{
						Bucket:    aws.String(bucketName),
						Key:       marker.Key,
						VersionId: marker.VersionId,
					}
					_, err := svc.DeleteObject(context.TODO(), deleteObjectInput)
					if err != nil {
						log.Printf("nao deu pra excluir o delete marker '%s', vou fingir q nao vi: %v", *marker.Key, err)
					} else {
						fmt.Printf("delete marker for %s (%s) foi cinzado\n", *marker.Key, *marker.VersionId)
					}

					waitGroup.Done()
					counter++
					fmt.Printf("matamo '%d' coisas até agora\n", counter)
				}(marker)
			}

			waitGroup.Wait()
		}
	}

	abortMultipartUploads := func() {
		listMultipartUploadsInput := &s3.ListMultipartUploadsInput{
			Bucket: aws.String(bucketName),
		}

		multipartPaginator := s3.NewListMultipartUploadsPaginator(svc, listMultipartUploadsInput)

		for multipartPaginator.HasMorePages() {
			page, err := multipartPaginator.NextPage(context.TODO())
			if err != nil {
				log.Fatalf("deu merda na listagem dos uploads multipart, %v", err)
			}

			waitGroup := sync.WaitGroup{}
			waitGroup.Add(len(page.Uploads))

			for _, upload := range page.Uploads {
				go func(upload types.MultipartUpload) {
					abortMultipartUploadInput := &s3.AbortMultipartUploadInput{
						Bucket:   aws.String(bucketName),
						Key:      upload.Key,
						UploadId: upload.UploadId,
					}
					_, err := svc.AbortMultipartUpload(context.TODO(), abortMultipartUploadInput)
					if err != nil {
						log.Printf("ninguém liga pro err no upload '%s': %v", *upload.Key, err)
					} else {
						fmt.Printf("head shot no upload %s\n", *upload.Key)
					}
					waitGroup.Done()
				}(upload)
			}

			waitGroup.Wait()
		}
	}

	deleteObjectVersions()

	abortMultipartUploads()

	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err = svc.DeleteBucket(context.TODO(), deleteBucketInput)
	if err != nil {
		log.Fatalf("ui ui ui, num pode jogar o bucket fora: %v", err)
	}

	fmt.Printf("bucket '%s' foi pro kar@!40, chefia! agora sim nao vai subir ninguém\n", bucketName)
}
