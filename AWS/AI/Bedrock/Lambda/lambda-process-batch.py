import json
import boto3
import base64
import io
import logging
from PIL import Image
from botocore.exceptions import ClientError
import uuid

brt = boto3.client(service_name='bedrock-runtime', region_name='us-east-1')
s3 = boto3.client('s3', region_name='ap-northeast-2')
dynamodb = boto3.client('dynamodb', region_name='ap-northeast-2')

# Custom exception for image generation errors
class ImageError(Exception):
    def __init__(self, message):
        self.message = message

# Set up logging
logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

def generate_image(model_id, body):
    logger.info(
        "Generating image with Amazon Titan Image Generator G1 model %s", model_id)

    accept = "application/json"
    content_type = "application/json"

    response = brt.invoke_model(
        body=body, modelId=model_id, accept=accept, contentType=content_type
    )
    response_body = json.loads(response.get("body").read())

    base64_image = response_body.get("images")[0]
    base64_bytes = base64_image.encode('ascii')
    image_bytes = base64.b64decode(base64_bytes)

    finish_reason = response_body.get("error")

    if finish_reason is not None:
        raise ImageError(f"Image generation error. Error is {finish_reason}")

    logger.info(
        "Successfully generated image with Amazon Titan Image Generator G1 model %s", model_id)

    return image_bytes

def lambda_handler(event, context):
    # Entrypoint for Amazon Titan Image Generator G1 example.
    logging.basicConfig(level=logging.INFO,
                        format="%(levelname)s: %(message)s")

    model_id = '<model-id>'
    s3_bucket = '<bucket-name>'
    output_prefix_dir = '<bucket-prefix-directory>'
    table_name = '<dynamo-db-table-name>'

    results = []

    for prompt in event:
        unique_filename = str(uuid.uuid4())
        
        prompt_text = f"Create an image that combines the three keywords [{prompt['keyword1']}({prompt['category1']}), {prompt['keyword2']}({prompt['category2']}), {prompt['keyword3']}({prompt['category3']})] into a single entity, drawing it as if a artist poured their soul into the artwork."

        body = json.dumps({
            "taskType": "TEXT_IMAGE",
            "textToImageParams": {
                "text": prompt_text
            },
            "imageGenerationConfig": {
                "numberOfImages": 1,
                "height": 512,
                "width": 512,
                "cfgScale": 9.5,
                "seed": 0
            }
        })

        try:
            image_bytes = generate_image(model_id=model_id, body=body)
            
            # Save image to /tmp directory
            image = Image.open(io.BytesIO(image_bytes))
            output_filename = f"{unique_filename}.png"
            image_path = f'/tmp/{output_filename}'
            image.save(image_path)

            # Upload image to S3
            s3_key = f'{output_prefix_dir}{output_filename}'
            s3.upload_file(image_path, s3_bucket, s3_key)

            # for image identifier
            category_id1, category_id2, category_id3 = prompt['category_id1'], prompt['category_id2'], prompt['category_id3']
            tag_id1, tag_id2, tag_id3 = prompt['tag_id1'], prompt['tag_id2'], prompt['tag_id3']

            category1_id = category_id1 * 100 + tag_id1
            category2_id = category_id2 * 100 + tag_id2
            category3_id = category_id3 * 100 + tag_id3

            current_image_id = str(category1_id) + str(category2_id) + str(category3_id)

            # Save metadata to DynamoDB
            dynamodb.put_item(
                TableName=table_name,
                Item={
                    'image_id': {'N': str(current_image_id)},
                    'image_URL': {'S': f's3://{s3_bucket}/{s3_key}'},
                    f"category{category_id1}_tag_id": {'S': str(tag_id1)},
                    f"category{category_id2}_tag_id": {'S': str(tag_id2)},
                    f"category{category_id3}_tag_id": {'S': str(tag_id3)}
                }
            )

            logger.info(f"Image successfully generated and uploaded to s3://{s3_bucket}/{s3_key}")

            results.append({
                'prompt': prompt_text,
                'image_URL': f's3://{s3_bucket}/{s3_key}',
                'image_id': int(current_image_id)
            })

        except ClientError as err:
            message = err.response["Error"]["Message"]
            logger.error("A client error occurred: %s", message)
            return {
                'statusCode': 400,
                'body': json.dumps(f"A client error occurred: {message}")
            }
        except ImageError as err:
            logger.error(err.message)
            return {
                'statusCode': 500,
                'body': json.dumps(err.message)
            }
        except Exception as e:
            logger.error(f"An error occurred: {str(e)}")
            return {
                'statusCode': 500,
                'body': json.dumps(f"An error occurred: {str(e)}")
            }

    return {
        'statusCode': 200,
        'body': json.dumps(results)
    }
