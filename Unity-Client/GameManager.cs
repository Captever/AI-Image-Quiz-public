using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Events;
using UnityEngine.SceneManagement;

public enum GameState
{
    None,
    InLobby,
    InRoom,
    InGame
}

public class GameManager : SingletonObj<GameManager>
{
    // public variable
    [HideInInspector] public string PlayerNickname { get; private set; }
    [HideInInspector] public WebSocketClient WebSocketClient { get; private set; }
    [HideInInspector] public SelectionManager SelectionManager { get; private set; }
    [HideInInspector] public ChatManager ChatManager { get; private set; }

    // # UnityEvent - delegate handler replacement
    // ## UI state
    public UnityEvent<GameState> OnStateChanged = new UnityEvent<GameState>();

    // ## Lobby & Room
    public UnityEvent<Lobby> OnUpdatedRoomList = new UnityEvent<Lobby>();
    public UnityEvent<Room> OnUpdatedParticipantsList = new UnityEvent<Room>();
    public UnityEvent<int> OnUpdatedStartCountdown = new UnityEvent<int>();
    public UnityEvent OnEnabledStart = new UnityEvent();

    // ## Game
    public UnityEvent<Texture2D> OnStartedGame = new UnityEvent<Texture2D>();
    public UnityEvent<int> OnUpdatedQuizCountdown = new UnityEvent<int>();
    public UnityEvent<int> OnShownQuizImage = new UnityEvent<int>();
    public UnityEvent<List<string>> OnStartedQuiz = new UnityEvent<List<string>>();
    public UnityEvent<int> OnUpdatedQuizTimer = new UnityEvent<int>();
    public UnityEvent<List<string>> OnRevealedAnswer = new UnityEvent<List<string>>();
    public UnityEvent<string, string, string> OnOccuredCorrectAnswer = new UnityEvent<string, string, string>();
    public UnityEvent<Dictionary<string, int>> OnEndedGame = new UnityEvent<Dictionary<string, int>>();

    private ConcurrentQueue<Action> mainThreadActions = new ConcurrentQueue<Action>();

    private void Awake()
    {
        WebSocketClient = GetComponent<WebSocketClient>();
        if (WebSocketClient == null)
        {
            WebSocketClient = gameObject.AddComponent<WebSocketClient>();
        }
        SelectionManager = GetComponent<SelectionManager>();
        if (SelectionManager == null)
        {
            SelectionManager = gameObject.AddComponent<SelectionManager>();
        }
    }
    private void Update()
    {
        // 메인 스레드에서 실행할 작업을 처리
        while (mainThreadActions.TryDequeue(out var action))
        {
            action?.Invoke();
        }
    }

    public void EnqueueMainThreadAction(Action action)
    {
        mainThreadActions.Enqueue(action);
    }


    // About GameState
    public GameState CurrentState { get; private set; }
    public void SetState(GameState newState)
    {
        if (CurrentState != newState)
        {
            CurrentState = newState;
            OnStateChanged.Invoke(newState);
        }
    }

    public void SetCurrentRoomList(Lobby lobby)
    {
        OnUpdatedRoomList.Invoke(lobby);
    }

    public void SetCurrentParticipantsList(Room room)
    {
        OnUpdatedParticipantsList.Invoke(room);
    }

    public void SetScriptChatManager(ChatManager chatManager)
    {
        ChatManager = chatManager;
    }

    public void UpdateStartCountdown(int countdown)
    {
        OnUpdatedStartCountdown.Invoke(countdown);
    }
    public void EnableStart()
    {
        OnEnabledStart.Invoke();
    }
    public void StartGame(byte[] spriteSheetBytes)
    {
        SetState(GameState.InGame);

        // 다운로드 받은 SpriteSheetBytes를 Texture2D로 만들어서 현재 게임에 적용
        Texture2D spriteSheet = new Texture2D(1536, 1536);
        spriteSheet.LoadImage(spriteSheetBytes);

        OnStartedGame.Invoke(spriteSheet);
    }
    public void UpdateQuizCountdown(int countdown)
    {
        OnUpdatedQuizCountdown.Invoke(countdown);
    }
    public void ShowQuizImage(int currentQuizIndex)
    {
        OnShownQuizImage.Invoke(currentQuizIndex);
    }
    public void StartQuiz(List<string> categoryNames)
    {
        OnStartedQuiz.Invoke(categoryNames);
    }
    public void UpdateQuizRemainingTime(int remainingTime)
    {
        OnUpdatedQuizTimer.Invoke(remainingTime);
    }
    public void RevealAnswer(List<string> keywords)
    {
        OnRevealedAnswer.Invoke(keywords);
    }
    public void OccurCorrectAnswer(string clientNickname, string category_name, string keyword)
    {
        OnOccuredCorrectAnswer.Invoke(clientNickname, category_name, keyword);
    }
    public void EndGame(Dictionary<string, int> scores)
    {
        OnEndedGame.Invoke(scores);
    }



    // # Init Scene
    // ## Player Nickname
    public void SetNickname(string nickname)
    {
        // TODO: 닉네임 조건 확인하기(비속어 등)
        PlayerNickname = nickname;
        Debug.Log("닉네임이 '" + PlayerNickname + "'으로 설정되었습니다.");
    }

    // # Function Plug-in
    public void LoadNextScene()
    {
        SceneManager.LoadScene(SceneManager.GetActiveScene().buildIndex + 1);
    }
    public static void RemoveAllChildren(GameObject parent)
    {
        // 자식 객체의 수를 미리 저장합니다.
        int childCount = parent.transform.childCount;

        // 모든 자식 객체를 반복하여 삭제합니다.
        for (int i = 0; i < childCount; i++)
        {
            Destroy(parent.transform.GetChild(i).gameObject);
        }
    }
}
