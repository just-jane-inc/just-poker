import requests
import json
import uuid

base_url= "http://localhost:7653"

def create_user():
    url = f'{base_url}/user'
    payload = {
        'username': str(uuid.uuid4()),
        'twitch_id': "some-id-2",
    }

    return requests.post(url, json=payload)

def delete_me(token: str):
    print(f'delete: {token}')
    url = f'{base_url}/user/me'
    header = {
        'Authorization': f'Bearer {token}',
        'Content-Type': 'application/json'
    }

    return requests.delete(url, headers=header)

def cycle_key(userid: str):
    url = f'{base_url}/user/{userid}/key'
    return requests.delete(url)

def get_users(twitch_id: str):
    url = f'{base_url}/user/twitch/{twitch_id}'
    return requests.get(url)

def test_one():
    create_user_response = create_user()
    assert create_user_response.status_code == 200

    key = create_user_response.json()['data']
    resp = cycle_key(key['user_id'])
    assert resp.status_code == 200

    token = resp.json()['data']['token']
    assert delete_me(token).status_code == 200

if __name__ == "__main__":
    r = get_users("some-id-2")
    print(json.dumps(r.json(), indent=2))
