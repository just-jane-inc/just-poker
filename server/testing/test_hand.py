import requests
import json
import argparse
import pytest
from typing import Dict, List

parser = argparse.ArgumentParser(description='sure')
parser.add_argument('--new-user')
parser.add_argument('--get-user')
parser.add_argument('--state')
parser.add_argument('--hand', nargs='+')
parser.add_argument('--all-check', action='store_true')
parser.add_argument('--one-raise', action='store_true')
parser.add_argument('--player-fold', action='store_true')
parser.add_argument('--all-in', action='store_true')
parser.add_argument('--just-start', action='store_true')
parser.add_argument('--test-all', action='store_true')

base_url= "http://localhost:7653"

def create_new_game(player_count: int, starting_chips: Dict[str, int]) -> str:
    url = f'{base_url}/game'
    payload = {
        'player_count': player_count,
        'starting_chips': starting_chips,
        'big_blind': 100,
        'small_blind': 50
    }

    response = requests.post(url, json=payload)
    if response.status_code == 200:
        return response.json()['data']
    else:
        print(response.status_code)
        print(response.json())
        return None

def get_auth_header(token: str):
    return {'Authorization': f'Bearer {token}', 'Content-Type': 'application/json'}

def join_game(game_id: str, player_id: str, token):
    url = f'{base_url}/game/{game_id}/player'
    return requests.post(url, headers=get_auth_header(token))

def start_game(game_id: str):
    url = f'{base_url}/game/{game_id}/started'
    response = requests.post(url)
    print(response.status_code)

def get_current_state(game_id: str) -> object:
    url = f'{base_url}/game/{game_id}/state'
    response = requests.get(url)
    print(response.status_code)
    return response.json()

def player_action(game_id: str, player_id: str, action: str, chips: Dict[str, int], just_status=True):
    url = f'{base_url}/game/{game_id}/action/'
    payload = {'player_id':player_id, 'intent': action, 'chips':chips}
    response = requests.post(url, json=payload)
    if just_status:
        return response.status_code
    return response

def test_player_fold():
    game_id = create_new_game(4, {"5":10, "20":100})
    join_game(game_id, '42')
    join_game(game_id, '420')
    join_game(game_id, '69')
    join_game(game_id, '67')
    start_game(game_id)

    # setup
    assert player_action(game_id, '69', 'ante', {'10':5}) == 200
    assert player_action(game_id, '67', 'ante', {'10':10}) == 200

    # pre-flop
    assert player_action(game_id, '42', 'call', {'10':10}) == 200
    assert player_action(game_id, '420', 'call', {'10':10}) == 200
    assert player_action(game_id, '69', 'call', {'10':5}) == 200 # to-call == 150
    assert player_action(game_id, '67', 'fold', {}) == 200

    # flop
    assert player_action(game_id, '69', 'check', {}) == 200
    assert player_action(game_id, '42', 'raise', {'10':15}) == 200
    assert player_action(game_id, '420', 'call', {'10':15}) == 200
    assert player_action(game_id, '69', 'call', {'10':15}) == 200

    # turn
    assert player_action(game_id, '69', 'check', {}) == 200
    assert player_action(game_id, '42', 'check', {}) == 200
    assert player_action(game_id, '420', 'check', {}) == 200

    # river
    assert player_action(game_id, '69', 'check', {}) == 200
    assert player_action(game_id, '42', 'check', {}) == 200
    assert player_action(game_id, '420', 'check', {}) == 200

def test_raise_once():
    game_id = create_new_game(4, {'10':20, '50':5, '100':2})
    users = []
    for i in range(4):
        name, twitch_id = (f"test-name-{i}", f"test-twitch-id-{i}")
        resp = create_user(name, twitch_id)
        assert resp.status_code == 200
        d = resp.json()['data']
        id, token = d['user_id'], d['token']
        user = (id, name, token)
        users.append(user)

        join_game(game_id, id, user[2])

    start_game(game_id)

    # setup
    assert player_action(game_id, users[2][0], 'ante', {'50':1}) == 200
    assert player_action(game_id, users[3][0], 'ante', {'100':1}) == 200

    # pre-flop
    assert player_action(game_id, users[0][0], 'call', {'10':10}) == 200
    assert player_action(game_id, users[1][0], 'call', {'50':2}) == 200
    assert player_action(game_id, users[2][0], 'raise', {'100':1}) # to-call == 150
    assert player_action(game_id, users[3][0], 'call', {'10':5}) == 200

    print(json.dumps(get_current_state(game_id), indent=2))

    assert player_action(game_id, users[0][0], 'call', {'10':5}) == 200
    assert player_action(game_id, users[1][0], 'call', {'10':5}) == 200

    # flop
    assert player_action(game_id, users[2][0], 'check', {}) == 200
    assert player_action(game_id, users[3][0], 'check', {}) == 200
    assert player_action(game_id, users[0][0], 'check', {}) == 200
    assert player_action(game_id, users[1][0], 'check', {}) == 200

    # turn
    assert player_action(game_id, users[2][0], 'check', {}) == 200
    assert player_action(game_id, users[3][0], 'check', {}) == 200
    assert player_action(game_id, users[0][0], 'check', {}) == 200
    assert player_action(game_id, users[1][0], 'check', {}) == 200

    # river
    assert player_action(game_id, users[2][0], 'check', {}) == 200
    assert player_action(game_id, users[3][0], 'check', {}) == 200
    assert player_action(game_id, users[0][0], 'check', {}) == 200
    assert player_action(game_id, users[1][0], 'check', {}) == 200

    print(game_id)

def test_all_in():
    game_id = create_new_game(4, {"5":10, "20":100})
    users = []
    for i in range(4):
        name, twitch_id = (f"test-name-{i}", f"test-twitch-id-{i}")
        resp = create_user(name, twitch_id)
        assert resp.status_code == 200
        d = resp.json()['data']
        id, token = d['user_id'], d['token']
        user = (id, name, token)
        users.append(user)

        join_game(game_id, id, user[2])

    start_game(game_id)

    # setup
    assert player_action(game_id, users[2][0], 'ante', {'10':5}) == 200
    assert player_action(game_id, users[3][0], 'ante', {'10':10}) == 200

    # pre-flop
    assert player_action(game_id, users[0][0], 'all_in', {'10':10})

    state = get_current_state(game_id)
    print(json.dumps(state, indent=2))

    assert player_action(game_id, users[1][0], 'all_in', {'10':10}) == 200
    assert player_action(game_id, users[2][0], 'all_in', {'10':5}) == 200
    assert player_action(game_id, users[3][0], 'all_in', {}) == 200


def test_all_check():
    game_id = create_new_game(4, {"5":10, "20":100})

    join_game(game_id, '42') # "under the gun"
    join_game(game_id, '420') # button
    join_game(game_id, '69') # small_blind
    join_game(game_id, '67') # big blind
    start_game(game_id)

    # setup
    assert player_action(game_id, '69', 'ante', {'10':5}) == 200
    assert player_action(game_id, '67', 'ante', {'10':10}) == 200

    # pre-flop
    assert player_action(game_id, '42', 'call', {'10':10}) == 200
    assert player_action(game_id, '420', 'call', {'10':10}) == 200
    assert player_action(game_id, '69', 'call', {'10':5}) == 200
    assert player_action(game_id, '67', 'check', {}) == 200

    # flop
    assert player_action(game_id, '69', 'check', {}) == 200
    assert player_action(game_id, '67', 'check', {}) == 200
    assert player_action(game_id, '42', 'check', {}) == 200
    assert player_action(game_id, '420', 'check', {}) == 200

    # turn
    assert player_action(game_id, '69', 'check', {}) == 200
    assert player_action(game_id, '67', 'check', {}) == 200
    assert player_action(game_id, '42', 'check', {}) == 200
    assert player_action(game_id, '420', 'check', {}) == 200

    # river
    assert player_action(game_id, '69', 'check', {}) == 200
    assert player_action(game_id, '67', 'check', {}) == 200
    assert player_action(game_id, '42', 'check', {}) == 200
    assert player_action(game_id, '420', 'check', {}) == 200

    print(game_id)

def test_hand_eval():
    assert EvaluateHand(['As', 'Ac', 'Ah', 'Ad', 'Ks']) == 11
    assert EvaluateHand(['As', 'Ks', 'Qs', 'Js', 'Ts']) == 1
    assert EvaluateHand(['As', 'As', 'Qs', 'Js', 'Ts']) == -1

def GetUser(user_id: str):
    url = f'{base_url}/user/{user_id}'

    response = requests.get(url)
    if response.status_code == 200:
        return response.json()['data']['display_name']
    else:
        print(response.status_code)
        print(response.json())
        return ""

def create_user(username: str, twitch_id: str, usertype: str = 'test'):
    url = f'{base_url}/user'
    payload = {
        'username': username,
        'user_type': usertype,
        'twitch_id': twitch_id
    }

    return requests.post(url, json=payload)

def EvaluateHand(hand: List[str]):
    url = f'{base_url}/utility/hand-eval'
    payload = [{'rank': ord(c[0]), 'suit': ord(c[1])} for c in hand]
    print(json.dumps(payload))
    response = requests.post(url, json=payload)
    if response.status_code != 200:
        return -1

    return response.json()['data']

if __name__ == '__main__':
    args = parser.parse_args()

    if (args.new_user):
        resp = create_user(args.new_user)
        print(json.dumps(resp, indent=2))

    if (args.test_all):
        test_all_check()
        test_all_in()
        test_raise_once()
        test_player_fold()

    if (args.all_check):
        test_all_check()

    if (args.all_in):
        test_all_in()

    if (args.one_raise):
        test_raise_once()

    if (args.player_fold):
        test_player_fold()

    if (args.hand):
        EvaluateHand(args.hand)

    if (args.get_user):
        GetUser(args.get_user)
