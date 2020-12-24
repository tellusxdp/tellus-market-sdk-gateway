import requests


def get_tool_token(token):
    resp = requests.post(
        "https://sdk.tellusxdp.com/api/manager/v2/auth/token/",
        json={"product_id": "2f59c093-4e80-419a-8584-dcee6589d3d2"},
        headers={"Authorization": "Bearer {}".format(token)}
    )
    resp.raise_for_status()
    data = resp.json()
    return data["token"]


if __name__ == "__main__":
    token = "2836253f-42b2-4e6d-8b5d-d2234031b6bf"
    tool_token = get_tool_token(token)

    url = "http://127.0.0.1:8000/demo/wif82af3w39s/api/kaminari_a.php?lat=35.099&lon=138.631&dtime=20190826202000&radi=1.0&format=json"
    resp = requests.get(
        url, headers={"Authorization": "Bearer {}".format(tool_token)})
    resp.raise_for_status()
    print(resp.text)
