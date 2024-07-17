import json

def lambda_handler(event, context):
    prompts = event['prompts']
    batch_size = event['batch_size']
    
    # Split prompts into batches of batch_size
    batches = [prompts[i:i + batch_size] for i in range(0, len(prompts), batch_size)]
    
    # Return the batches wrapped in a dictionary
    return {
        'batches': batches
    }