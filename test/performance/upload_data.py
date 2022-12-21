"""
Python Module to interact with aws opensearch instance
"""

import os
import string
import random
from datetime import datetime

import yaml
import boto3
from opensearchpy import OpenSearch, RequestsHttpConnection, AWSV4SignerAuth


def create_index(os_client, index_name, number_of_shards=4):
    """ Method to create index in opensearch instance """
    index_body = {
      'settings': {
        'index': {
          'number_of_shards': number_of_shards
        }
      }
    }
    response = os_client.indices.create(index_name, body=index_body)
    print('\nCreating index:')
    print(response)


def delete_index(os_client, index_name):
    """ Method to delete index from opensearch instance """
    response = os_client.indices.delete(
      index=index_name
    )
    print('\nDeleting index:')
    print(response)


def add_document_to_index(os_client, index_name, doc_id, document):
    """ Add a document to the index """
    response = os_client.index(
        index=index_name,
        body=document,
        id=doc_id,
        refresh=True
    )
    print('\nAdding document:')
    print(response)


def delete_a_document(os_client, index_name, doc_id):
    """ Delete a document from index """
    response = os_client.delete(
        index=index_name,
        id=doc_id
    )
    print('\nDeleting document:')
    print(response)


def search_document(os_client, index_name):
    """ Sample search for the document """
    qval = 'miller'
    query = {
      'size': 5,
      'query': {
        'multi_match': {
          'query': qval,
          'fields': ['title^2', 'director']
        }
      }
    }
    response = os_client.search(
        body=query,
        index=index_name
    )
    print('\nSearch results:')
    print(response)


def setup_os_client():
    """ Setup the open search client """
    host = os.environ['OS_HOST']  # cluster endpoint, for ex: my-domain.us-east-1.es.amazonaws.com
    region = os.environ['OS_REGION']
    credentials = boto3.Session().get_credentials()
    auth = AWSV4SignerAuth(credentials, region)

    os_client = OpenSearch(
        hosts=[{'host': host, 'port': 443}],
        http_auth=auth,
        use_ssl=True,
        verify_certs=True,
        connection_class=RequestsHttpConnection
    )
    return os_client


def read_metric_data(file_name):
    """ Read from the metrics file and construct the document """
    with open(file_name, encoding="utf-8") as file_d:
        content = yaml.load(file_d, Loader=yaml.FullLoader)
        kpi_data = content['kpi']
    for data in kpi_data:
        if data['name'] == 'usage':
            for metric in data['metrics']:
                if metric['name'] == 'Memory_MiB':
                    memory_average = metric['average']
                    memory_maximum = metric['maximum']
                elif metric['name'] == 'CPU_millicores':
                    cpu_average = metric['average']
                    cpu_maximum = metric['maximum']
    dt_string = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    return {'upload_date': dt_string,
            'memory_average': float(memory_average),
            'memory_maximum': float(memory_maximum),
            'cpu_average': float(cpu_average),
            'cpu_maximum': float(cpu_maximum),
            'memory_average_threshold': get_average_threshold_mem(),
            'memory_maximum_threshold': get_maximum_threshold_mem(),
            'cpu_average_threshold': get_average_threshold_cpu(),
            'cpu_maximum_threshold': get_maximum_threshold_cpu()}


def generate_id():
    """ Generate a random id of length 6 """
    length = 6
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


def get_average_threshold_mem():
    """ returns current threshold value 150 """
    return float(os.environ['TEST_PERFORMANCE_AVG_MEMORY'])


def get_maximum_threshold_mem():
    """ returns current threshold value 200 """
    return float(os.environ['TEST_PERFORMANCE_MAX_MEMORY'])


def get_average_threshold_cpu():
    """ returns current thresholds value 20 """
    return float(os.environ['TEST_PERFORMANCE_AVG_CPU'])


def get_maximum_threshold_cpu():
    """ returns current thresholds value 100 """
    return float(os.environ['TEST_PERFORMANCE_MAX_CPU'])


if __name__ == '__main__':
    OS_INDEX_NAME = 'sbo-perf-data'
    client = setup_os_client()
    # create_index(client, index_name)
    # delete_index(client, index_name)
    # delete_a_document(client, index_name, id)

    metric_file_name = os.environ['KPI_YAML_FILE']
    doc = read_metric_data(metric_file_name)
    # doc = {'upload_date': '2022-12-14 06:30:30',
    # 'memory_average' : 68.2,
    # 'memory_maximum': 98.2,
    # 'cpu_average': 10.5,
    # 'cpu_maximum': 90.2,
    # 'memory_average_threshold': 150,
    # 'memory_maximum_threshold': 200,
    # 'cpu_average_threshold': 20,
    # 'cpu_maximum_threshold': 100}
    RANDOM_DOC_ID = generate_id()
    print(f"Random Generated ID: {RANDOM_DOC_ID}")
    add_document_to_index(client, OS_INDEX_NAME, RANDOM_DOC_ID, doc)
