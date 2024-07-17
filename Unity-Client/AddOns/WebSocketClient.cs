using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System;
using System.Collections;
using System.Collections.Concurrent;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.Networking;
using WebSocketSharp;

public class WebSocketClient : MonoBehaviour
{
    private WebSocket ws;
    [HideInInspector] public string ClientUUID { get; private set; }
    [HideInInspector] public string RoomUUID { get; private set; }

    public string serverUrl = Config.ServerUrl;

    public void ConnectToServer()
    {
        ws = new WebSocket($"ws://{serverUrl}/ws");

        ws.OnOpen += (sender, e) =>
        {
            Debug.Log("WebSocket connection opened");
        };

        ws.OnMessage += (sender, e) =>
        {
            Debug.Log("Message received: " + e.Data);

            // Handle incoming message
            HandleServerMessage(e.Data);
        };

        ws.OnError += (sender, e) =>
        {
            Debug.LogError("WebSocket error: " + e.Message);
        };

        ws.OnClose += (sender, e) =>
        {
            Debug.Log("WebSocket connection closed.");
        };

        ws.Connect();
    }

    public void ConnectToMaster()
    {
        SendDataToServer("ConnectToMaster", new ConnectionMessage(GameManager.Instance.PlayerNickname));
    }
    public void CreateRoom(string roomName, int maxParticipants)
    {
        SendDataToServer("CreateRoom", new RoomMessage(roomName, "", maxParticipants));
    }
    public void JoinRoom(string roomUUID)
    {
        SendDataToServer("JoinRoom", new RoomMessage("", roomUUID, 0));
    }
    public void LeftRoom()
    {
        SendDataToServer("LeftRoom", new RoomMessage("", RoomUUID, 0));
        RoomUUID = "";
        // TODO: 방이 정상적으로 잘 나가졌는지 확인할 수단 마련
        //     + 잘 나가졌을 때 UI가 전환되도록
        GameManager.Instance.SetState(GameState.InLobby);
    }
    public void StartGame()
    {
        RoomMessage roomMessage = new RoomMessage("", RoomUUID, 0);
        SendDataToServer("StartGame", roomMessage);
    }
    public void SendChatMessage(string content)
    {
        SendDataToServer("SendChatMessage", new ChatMessage(RoomUUID, content));
    }


    private void SendDataToServer(string action, object jsonData)
    {
        if (ws.ReadyState == WebSocketState.Open)
        {
            var jsonMessage = new MessageWrapper(action, ClientUUID, jsonData).ToJson();
            ws.Send(jsonMessage);
        }
        else
        {
            Debug.LogWarning("WebSocket is not open. Data not sent: " + action + " | " + jsonData);
        }
    }

    void HandleServerMessage(string jsonMessage)
    {
        MessageWrapper smw = new MessageWrapper(jsonMessage);

        try
        {
            JObject parsedMessage = null;

            if (smw.Data != null)
            {
                parsedMessage = JObject.FromObject(smw.Data);
            }
            else
            {
                Debug.Log("smw.Data is null");
            }

            switch (smw.Action)
            {
                case "OnConnectedToMaster":
                    OnConnectedToMaster(parsedMessage);
                    break;
                case "OnJoinedLobby":
                    OnJoinedLobby(parsedMessage);
                    break;
                case "OnUpdatedLobby":
                    OnUpdatedLobby(parsedMessage);
                    break;
                case "OnCreatedRoom":
                    OnCreatedRoom(parsedMessage);
                    break;
                case "OnJoinedRoom":
                    OnJoinedRoom(parsedMessage);
                    break;
                case "OnUpdatedRoom":
                    OnUpdatedRoom(parsedMessage);
                    break;
                case "OnRecievedChatMessage":
                    OnRecievedChatMessage(parsedMessage);
                    break;
                case "OnRecievedSystemMessage":
                    OnRecievedSystemMessage(parsedMessage);
                    break;
                case "OnUpdatedStartCountdown":
                    OnUpdatedStartCountdown(parsedMessage);
                    break;
                case "OnCancelledStartCountdown":
                    OnCancelledStartCountdown();
                    break;
                case "OnEnabledStart":
                    OnEnabledStart();
                    break;
                case "OnStartedGame":
                    OnStartedGame(parsedMessage);
                    break;
                case "OnStartedQuiz":
                    OnStartedQuiz(parsedMessage);
                    break;
                case "OnUpdatedQuizCountdown":
                    OnUpdatedQuizCountdown(parsedMessage);
                    break;
                case "OnShownQuizImage":
                    OnShownQuizImage(parsedMessage);
                    break;
                case "OnUpdatedQuizRemainingTime":
                    OnUpdatedQuizRemainingTime(parsedMessage);
                    break;
                case "OnRevealedAnswer":
                    OnRevealedAnswer(parsedMessage);
                    break;
                case "OnOccuredCorrectAnswer":
                    OnOccuredCorrectAnswer(parsedMessage);
                    break;
                case "OnEndedGame":
                    OnEndedGame(parsedMessage);
                    break;
                default:
                    Debug.LogWarning("Unknown action from server message: " + smw.Action);
                    break;
            }
        }
        catch (Exception ex)
        {
            Debug.LogError("Failed to parse JSON message: " + ex.Message);
        }
    }

    void OnConnectedToMaster(JObject parsedMessage)
    {
        string clientUUID = parsedMessage["clientUUID"].ToString();
        ClientUUID = clientUUID; // 변수에 저장
    }
    void OnJoinedLobby(JObject parsedMessage)
    {
        // TODO: 방 목록 불러오기
    }
    void OnUpdatedLobby(JObject sData)
    {
        Lobby lobby = sData.ToObject<Lobby>();

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.SetCurrentRoomList(lobby);
        });
    }
    void OnCreatedRoom(JObject parsedMessage)
    {
        string roomUUID = parsedMessage["roomUUID"].ToString();
        RoomUUID = roomUUID; // 변수에 저장
    }
    void OnJoinedRoom(JObject parsedMessage)
    {
        string roomUUID = parsedMessage["roomUUID"].ToString();
        RoomUUID = roomUUID;

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.SetState(GameState.InRoom);
        });
    }
    void OnUpdatedRoom(JObject sData)
    {
        Room room = sData.ToObject<Room>();

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.SetCurrentParticipantsList(room);
        });
    }
    void OnRecievedChatMessage(JObject sData)
    {
        string newMessage = sData["message"].ToString();

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.ChatManager.AddNewChatMessage(newMessage);
        });
    }
    void OnRecievedSystemMessage(JObject sData)
    {
        string newMessage = sData["message"].ToString();

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.ChatManager.AddNewSystemMessage(newMessage);
        });
    }
    void OnUpdatedStartCountdown(JObject sData)
    {
        int countdown = int.Parse(sData["startCountdownTime"].ToString());

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.UpdateStartCountdown(countdown);
        });
    }
    private void OnCancelledStartCountdown()
    {
        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.UpdateStartCountdown(0);
        });
    }
    private void OnEnabledStart()
    {
        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.EnableStart();
        });
    }
    private void OnStartedGame(JObject parsedMessage)
    {
        string spriteSheetBase64 = parsedMessage["spriteSheet"].ToString();
        byte[] spriteSheetBytes = Convert.FromBase64String(spriteSheetBase64);

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.StartGame(spriteSheetBytes);
        });
    }
    private void OnStartedQuiz(JObject parsedMessage)
    {
        try
        {
            List<string> categoryNames = parsedMessage["categoryNames"].ToObject<List<string>>();

            GameManager.Instance.EnqueueMainThreadAction(() =>
            {
                GameManager.Instance.StartQuiz(categoryNames);
            });
        }
        catch (Exception ex)
        {
            Debug.LogError("Error processing OnStartedQuiz: " + ex.Message);
        }
    }
    private void OnUpdatedQuizCountdown(JObject parsedMessage)
    {
        int countdown = int.Parse(parsedMessage["quizCountdownTime"].ToString());

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.UpdateQuizCountdown(countdown);
        });
    }
    private void OnShownQuizImage(JObject parsedMessage)
    {
        int currentQuizIndex = int.Parse(parsedMessage["currentQuizIndex"].ToString());

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.ShowQuizImage(currentQuizIndex);
        });
    }
    private void OnUpdatedQuizRemainingTime(JObject parsedMessage)
    {
        int remainingTime = int.Parse(parsedMessage["remainingTime"].ToString());
        
        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.UpdateQuizRemainingTime(remainingTime);
        });
    }
    private void OnRevealedAnswer(JObject parsedMessage)
    {
        try
        {
            // 'keywords' 필드를 추출하여 List<string> 형식으로 역직렬화
            List<string> keywords = parsedMessage["keywords"].ToObject<List<string>>();

            GameManager.Instance.EnqueueMainThreadAction(() =>
            {
                GameManager.Instance.RevealAnswer(keywords);
            });
        }
        catch (JsonException ex)
        {
            Debug.LogError("Failed to deserialize keywords: " + ex.Message);
        }
    }
    private void OnOccuredCorrectAnswer(JObject parsedMessage)
    {
        // TODO: 서버에서 정답자 발생 시 모든 클라이언트에 정답자와 정답을 뿌리고 있는데,
        // 정답자 발생 시스템 메시지를 띄울 수 있도록 할 것
        string clientNickname = parsedMessage["clientNickname"].ToString();
        string category_name = parsedMessage["category_name"].ToString();
        string keyword = parsedMessage["keyword"].ToString();

        // 정답 처리 (UI 업데이트 등)
        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.OccurCorrectAnswer(clientNickname, category_name, keyword);
        });
    }
    private void OnEndedGame(JObject parsedMessage)
    {
        Dictionary<string, int> scores = parsedMessage.ToObject<Dictionary<string, int>>();

        GameManager.Instance.EnqueueMainThreadAction(() =>
        {
            GameManager.Instance.EndGame(scores);
        });
    }


    void OnDestroy()
    {
        if (ws != null)
        {
            // TODO: session 부분이 제대로 구현된다면 clientUUID 사실상 필요 없을 듯
            SignoutSession();

            ws.Close();
            ws = null;
        }
    }
    //public void SendSignoutRequest(string sessionUUID, string clientUUID, string roomUUID)
    //{
    //    SendDataToServer("Signout", new SessionMessage(sessionUUID, clientUUID, roomUUID));
    //}
    void SignoutSession()
    {
        string sessionUUID = PlayerPrefs.GetString("SessionUUID", "");
        string clientUUID = PlayerPrefs.GetString("ClientUUID", "");
        string roomUUID = PlayerPrefs.GetString("RoomUUID", "");
        if (!string.IsNullOrEmpty(sessionUUID))
        {
            StartCoroutine(SignoutCoroutine(sessionUUID, clientUUID, roomUUID));
        }
    }

    // TODO: 작동 안함
    // 세션 로그아웃(현재는 강제 종료 시 사용)
    private IEnumerator SignoutCoroutine(string sessionUUID, string clientUUID, string roomUUID)
    {
        string signoutUrl = "http://" + Config.ServerUrl + "/signout";
        Debug.Log("로그아웃 시도 중: " + signoutUrl);
        SignoutRequest signoutRequest = new SignoutRequest { SessionUUID = sessionUUID, ClientUUID = clientUUID, RoomUUID = roomUUID };
        string jsonData = JsonUtility.ToJson(signoutRequest);

        UnityWebRequest request = new UnityWebRequest(signoutUrl, "POST");
        byte[] bodyRaw = System.Text.Encoding.UTF8.GetBytes(jsonData);
        request.uploadHandler = new UploadHandlerRaw(bodyRaw);
        request.downloadHandler = new DownloadHandlerBuffer();
        request.SetRequestHeader("Content-Type", "application/json");

        yield return request.SendWebRequest();

        if (request.result == UnityWebRequest.Result.Success)
        {
            Debug.Log("로그아웃 성공");
        }
        else
        {
            Debug.LogError("Signout failed: " + request.error);
        }
    }
    [System.Serializable]
    private class SignoutRequest
    {
        public string SessionUUID;
        public string ClientUUID;
        public string RoomUUID;
    }
}
