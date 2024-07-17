using Newtonsoft.Json;
using System;
using System.Collections.Generic;
using UnityEngine;


[System.Serializable]
public class MessageWrapper
{
    [JsonProperty("action")]
    public string Action { get; set; }  // 어떤 메시지인지 식별할 수 있는 헤더
    [JsonProperty("clientUUID")]
    public string ClientUUID { get; set; } // 누가 보낸 메시지인지 식별하기 위함
    [JsonProperty("data")]
    public object Data { get; set; }  // 메시지의 실제 데이터

    [JsonConstructor]
    public MessageWrapper(string action, string clientUUID, object data)
    {
        Action = action;
        ClientUUID = clientUUID;
        Data = data;
    }
    public MessageWrapper(string json)
    {
        try
        {
            MessageWrapper mw = JsonConvert.DeserializeObject<MessageWrapper>(json);
            Action = mw.Action;
            ClientUUID = mw.ClientUUID;
            Data = mw.Data;
        }
        catch (Exception ex)
        {
            Debug.LogError("Error deserializing MessageWrapper: " + ex.Message);
        }
    }

    public string ToJson()
    {
        return JsonConvert.SerializeObject(this);
    }
}
[System.Serializable]
public class ChatMessage
{
    [JsonProperty("roomUUID")]
    public string RoomUUID {  get; set; }
    [JsonProperty("content")]
    public string Content {  get; set; }

    public ChatMessage(string roomUUID, string content)
    {
        RoomUUID = roomUUID;
        Content = content;
    }

    public string ToJson()
    {
        return JsonConvert.SerializeObject(this);
    }
}
// CreateRoom 등 방 접근 관련에 사용될 클래스
[System.Serializable]
public class RoomMessage
{
    /* TODO: 방을 만들 때 사용되는 roomName과 maxParticipants
             방에 입장할 때 사용되는 roomUUID
             이 두 개를 한 클래스에서 공존시키는 게 맞나?*/
    [JsonProperty("roomName")]
    public string RoomName {  get; set; }
    [JsonProperty("roomUUID")]
    public string RoomUUID {  get; set; }
    [JsonProperty("maxParticipants")]
    public int MaxParticipants {  get; set; }

    public RoomMessage(string roomName, string roomUUID, int maxParticipants)
    {
        RoomName = roomName;
        RoomUUID = roomUUID;
        MaxParticipants = maxParticipants;
    }

    public string ToJson()
    {
        return JsonConvert.SerializeObject(this);
    }
}
// ConnectToMaster에 사용될 클래스
[System.Serializable]
public class ConnectionMessage
{
    [JsonProperty("clientNickname")]
    public string ClientNickname {  get; set; }

    public ConnectionMessage(string clientNickname)
    {
        ClientNickname = clientNickname;
    }

    public string ToJson()
    {
        return JsonConvert.SerializeObject(this);
    }
}
[System.Serializable]
public class SessionMessage
{
    [JsonProperty("sessionUUID")]
    public string SessionUUID { get; set; }
    [JsonProperty("clientUUID")]
    public string ClientUUID { get; set; }
    [JsonProperty("roomUUID")]
    public string RoomUUID { get; set; }

    public SessionMessage(string sessionUUID, string clientUUID, string roomUUID)
    {
        SessionUUID = sessionUUID;
        ClientUUID = clientUUID;
        RoomUUID = roomUUID;
    }

    public string ToJson()
    {
        return JsonConvert.SerializeObject(this);
    }
}


[System.Serializable]
public class Client
{
    [JsonProperty("clientNickname")]
    public string ClientNickname { get; set; }


    public bool Equals(Client other)
    {
        return other.ClientNickname.Equals(ClientNickname);
    }
}

[System.Serializable]
public class Room
{
    [JsonProperty("roomName")]
    public string RoomName { get; set; }

    [JsonProperty("maxParticipants")]
    public int MaxParticipants { get; set; }

    [JsonProperty("clients")]
    public Dictionary<string, Client> Clients { get; set; }

    [JsonProperty("masterClient")]
    public Client MasterClient { get; set; }

    [JsonProperty("roomState")]
    public RoomState RoomState { get; set; }
}

[System.Serializable]
public class RoomState
{
    [JsonProperty("isInGame")]
    public bool IsInGame { get; set; }

    [JsonProperty("isCountingDown")]
    public bool IsCountingDown { get; set; }

    [JsonProperty("countdownTime")]
    public int CountdownTime { get; set; }

    [JsonProperty("currentQuiz")]
    public Quiz CurrentQuiz { get; set; }

    [JsonProperty("quizzes")]
    public List<Quiz> Quizzes { get; set; }

    [JsonProperty("currentQuizIndex")]
    public int CurrentQuizIndex { get; set; }
}

[System.Serializable]
public class Quiz
{
    [JsonProperty("imageURL")]
    public string ImageURL { get; set; }

    [JsonProperty("keywords")]
    public List<string> Keywords { get; set; }

    [JsonProperty("guessedKeywords")]
    public List<string> GuessedKeywords { get; set; }

    [JsonProperty("remainingTime")]
    public int RemainingTime { get; set; }
}

[System.Serializable]
public class Lobby
{
    [JsonProperty("clients")]
    public Dictionary<string, Client> Clients { get; set; }

    [JsonProperty("rooms")]
    public Dictionary<string, Room> Rooms { get; set; }
}