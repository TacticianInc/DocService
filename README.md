# DocService

This is a MicroService written in Go. Its purpose is to have an API interface to S3 for the purposes of storing and retrieving raw files. To use the API see the dev folder where the Docker file is located. Uploading this to Amazon BeanStalk or to an EC2 or similar server using GIT, you can run it in a Docker container.

To make a call to the service, use the following:

TO SAVE:

POST: URL/doc/save/
JSON:
{
  "type":"mime type",
  "data":"base 64 binary file contents"
}

There is an option endpoint to get a file, however, it is not needed at the moment.
