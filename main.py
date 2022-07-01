#! /usr/bin/env python3
import json
import sys
import argparse

def read_json(myjson):
    try:
        json_object = json.loads(myjson)
    except ValueError as e:
        print("Invalid json")
        sys.exit(1)
    return json_object 

def json_parser(indict: dict, pre: list = None):
    pre = pre[:] if pre else []
    if isinstance(indict, dict):
        for key, value in indict.items():
            if isinstance(value, dict):
                for d in json_parser(value, pre + [key]):
                    yield d
            elif isinstance(value, list):
                for index, v in enumerate(value):
                    for d in json_parser(v, pre + [f'{key}[{index}]']):
                        yield d
            else:
                yield pre + [key, value]
    else:
        yield pre + [indict]

def create_output(list_of_entries: list) -> dict:
    output = {}
    for i in list_of_entries:
        value = i.pop(-1)
        key = ".".join(i)
        output[key]=value
    return output

def htmlifier(dict_of_flattened_json: dict) -> None:
    html = '<html><table border="1"><tr><th>Keys:</th><th>' + '</th><th>'.join(dict_of_flattened_json.keys()) 
    html += '</th></tr><tr><td>Values:</td><td>' + '</td><td>'.join(str(x) for x in dict_of_flattened_json.values())
    html += '</td></tr></table></html>'
    with open('result.html', 'w') as file:
        file.write(html)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('filename', nargs = '?', type = argparse.FileType('r'), default = sys.stdin)
    args = parser.parse_args()
    content = args.filename.read() 
    content = read_json(content)
    items_as_list = [*json_parser(content)]
    final_output = create_output(items_as_list)
    print(final_output)
    for key, value in final_output.items():
        print(f'{key} = {value}')
    htmlifier(create_output(items_as_list))   
    
