# image-resizer
image-resizer is an app that could resize your image and upload results to AWS S3 bucket  
  
To run an application it's necessary to specify env vars:  
*AWS_ACCESS_KEY_ID = {your id}*  
*AWS_SECRET_ACCESS_KEY = {your access key}*   
 
 And then run:  
*go run main.go -config config.json*  

 Api docs : [link](https://github.com/Dimitriy14/image-resizing/blob/master/api/swagger.yml)