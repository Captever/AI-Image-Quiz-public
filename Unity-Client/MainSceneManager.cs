using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class MainSceneManager : MonoBehaviour
{
    // # Chat System
    [SerializeField] private GameObject grpChatUI;

    // # In Lobby
    [SerializeField] private GameObject grpInLobby; // parent of pnlRoomList
    [SerializeField] private GameObject layRoomList;
    [SerializeField] private GameObject prefabRoomListItem;

    // ## Create Room
    [SerializeField] private Button btnCreateRoom;
    [SerializeField] private GameObject grpCreateRoom;
    [SerializeField] private Button btnExitCreateRoom;
    [SerializeField] private TMP_InputField inputRoomName;
    [SerializeField] private GameObject layMaxParticipants;
    [SerializeField] private GameObject prefabMaxParticipantsItem;
    [SerializeField] private Button btnConfirmCreateRoom;

    [SerializeField] private Button btnJoinRoom;

    // # In Room
    [SerializeField] private GameObject grpInRoom;
    [SerializeField] private GameObject layParticipantsList;
    [SerializeField] private GameObject prefabParticipantsListItem;

    [SerializeField] private TextMeshProUGUI txtStartCountdown;
    [SerializeField] private Button btnStartGame;
    [SerializeField] private Button btnLeftRoom;

    // # In Game
    [SerializeField] private GameObject grpInGame;

    [SerializeField] private TextMeshProUGUI txtQuizCountdown;

    private Texture2D currentQuizImageSpritesheet;
    [SerializeField] private Image imgQuiz;
    [SerializeField] private Texture2D defaultImg;
    [SerializeField] private TextMeshProUGUI txtQuizTimer;
    private List<string> gameAnswerCategoryNames; 
    [SerializeField] private List<TextMeshProUGUI> lblGameAnswers;
    [SerializeField] private List<TextMeshProUGUI> txtGameAnswers;

    [SerializeField] private GameObject grpScoreBoard;
    [SerializeField] private TextMeshProUGUI txtFinalScore;
    [SerializeField] private Button btnOkScoreBoard;

    // Start is called before the first frame update
    void Awake()
    {
        imgQuiz.sprite = Sprite.Create(
            defaultImg, new Rect(0, 0, defaultImg.width, defaultImg.height), new Vector2(0.5f, 0.5f));
        foreach (TextMeshProUGUI currTxtGameAnswer in txtGameAnswers)
        {
            currTxtGameAnswer.text = "_";
        }
        foreach (TextMeshProUGUI currLblGameAnswer in lblGameAnswers)
        {
            currLblGameAnswer.text = "?";
        }

        InitGroupCreateRoom();

        btnCreateRoom.onClick.AddListener(OnClickedCreateRoom);
        btnExitCreateRoom.onClick.AddListener(OnClickedExitCreateRoom);
        btnConfirmCreateRoom.onClick.AddListener(OnClickedConfirmCreateRoom);

        btnJoinRoom.onClick.AddListener(OnClickedJoinRoom);
        btnLeftRoom.onClick.AddListener(OnClickedLeftRoom);

        btnStartGame.interactable = false;
        btnStartGame.onClick.AddListener(OnClickedStartGame);
        txtStartCountdown.gameObject.SetActive(false);

        txtQuizTimer.text = "";
        txtFinalScore.text = "";

        grpScoreBoard.SetActive(false);
        btnOkScoreBoard.onClick.AddListener(OnClickedOkScoreBoard);
    }
    private void OnEnable()
    {
        // OnStateChanged 이벤트에 함수 추가
        GameManager.Instance.OnStateChanged.AddListener(HandleStateChanged);

        // TODO: InLobby 일 때만 방 목록 업데이트
        GameManager.Instance.OnUpdatedRoomList.AddListener(HandleUpdateRoomList);
        // TODO: InRoom 일 때만 참가자 목록 업데이트
        GameManager.Instance.OnUpdatedParticipantsList.AddListener(HandleUpdateParticipantsList);

        // Game 로직 관련 이벤트 조작
        GameManager.Instance.OnUpdatedStartCountdown.AddListener(HandleUpdateStartCountdown);
        GameManager.Instance.OnEnabledStart.AddListener(HandleEnableStart);
        GameManager.Instance.OnStartedGame.AddListener(HandleStartGame);
        GameManager.Instance.OnUpdatedQuizCountdown.AddListener(HandleUpdateQuizCountdown);
        GameManager.Instance.OnShownQuizImage.AddListener(HandleShowQuizImage);
        GameManager.Instance.OnStartedQuiz.AddListener(HandleStartQuiz);
        GameManager.Instance.OnUpdatedQuizTimer.AddListener(HandleUpdateQuizTimer);
        GameManager.Instance.OnRevealedAnswer.AddListener(HandleRevealQuizAnswer);
        GameManager.Instance.OnOccuredCorrectAnswer.AddListener(HandleOccurQuizAnswer);
        GameManager.Instance.OnEndedGame.AddListener(HandleEndGame);

        // 초기 상태를 로비로 설정
        GameManager.Instance.SetState(GameState.InLobby);
    }

    private void OnDisable()
    {
        // OnStateChanged 이벤트에 함수 제거
        GameManager.Instance.OnStateChanged.RemoveListener(HandleStateChanged);

        // TODO: InLobby 일 때만 방 목록 업데이트
        GameManager.Instance.OnUpdatedRoomList.RemoveListener(HandleUpdateRoomList);
        // TODO: InRoom 일 때만 참가자 목록 업데이트
        GameManager.Instance.OnUpdatedParticipantsList.RemoveListener(HandleUpdateParticipantsList);

        // Game 로직 관련 이벤트 조작
        GameManager.Instance.OnUpdatedStartCountdown.RemoveListener(HandleUpdateStartCountdown);
        GameManager.Instance.OnEnabledStart.RemoveListener(HandleEnableStart);
        GameManager.Instance.OnStartedGame.RemoveListener(HandleStartGame);
        GameManager.Instance.OnUpdatedQuizCountdown.RemoveListener(HandleUpdateQuizCountdown);
        GameManager.Instance.OnShownQuizImage.RemoveListener(HandleShowQuizImage);
        GameManager.Instance.OnStartedQuiz.AddListener(HandleStartQuiz);
        GameManager.Instance.OnUpdatedQuizTimer.RemoveListener(HandleUpdateQuizTimer);
        GameManager.Instance.OnRevealedAnswer.RemoveListener(HandleRevealQuizAnswer);
        GameManager.Instance.OnOccuredCorrectAnswer.RemoveListener(HandleOccurQuizAnswer);
        GameManager.Instance.OnEndedGame.RemoveListener(HandleEndGame);
    }


    private void HandleStateChanged(GameState newState)
    {
        grpInLobby.SetActive(newState == GameState.InLobby);
        grpInRoom.SetActive(newState == GameState.InRoom);
        grpInGame.SetActive(newState == GameState.InGame);
        grpChatUI.SetActive(newState != GameState.InLobby);

        //if (newState == GameState.InLobby)
        //{
        //    GameManager.Instance.OnUpdatedRoomList += HandleUpdateRoomList;
        //}
        //else
        //{
        //    GameManager.Instance.OnUpdatedRoomList -= HandleUpdateRoomList;
        //}
        //if (newState == GameState.InRoom)
        //{
        //    GameManager.Instance.OnUpdatedParticipantsList += HandleUpdateParticipantsList;
        //}
        //else
        //{
        //    GameManager.Instance.OnUpdatedParticipantsList -= HandleUpdateParticipantsList;
        //}
    }
    private void HandleUpdateRoomList(Lobby lobby)
    {
        GameManager.RemoveAllChildren(layRoomList);
        foreach (KeyValuePair<string, Room> roomKvp in lobby.Rooms)
        {
            GameObject roomListItem = Instantiate(prefabRoomListItem, layRoomList.transform);

            string roomUUID = roomKvp.Key;
            Room currRoom = roomKvp.Value;
            string roomName = currRoom.RoomName;
            string roomParticipantsCount = currRoom.Clients.Count.ToString();
            string roomMaxParticipants = currRoom.MaxParticipants.ToString();
            Client masterClient = currRoom.MasterClient;

            string roomTitle = roomName + " - 참가자 " + roomParticipantsCount + "/" + roomMaxParticipants;

            RoomListItem roomListItemComponent = roomListItem.GetComponent<RoomListItem>();
            roomListItemComponent.Setup(roomUUID, roomTitle, masterClient);

            // 게임 진행 중인 방의 색상 변경
            if (currRoom.RoomState.IsInGame)
            {
                roomListItemComponent.SetInGameColor();
            }
        }
    }
    private void HandleUpdateParticipantsList(Room room)
    {
        GameManager.RemoveAllChildren(layParticipantsList);
        foreach (KeyValuePair<string, Client> clientKvp in room.Clients)
        {
            GameObject participantsListItem = Instantiate(prefabParticipantsListItem, layParticipantsList.transform);

            string clientUUID = clientKvp.Key;
            Client currClient = clientKvp.Value;
            string clientName = currClient.ClientNickname;
            bool isMasterClient = false;
            if (room.MasterClient.Equals(currClient))
            {
                isMasterClient = true;
            }

            participantsListItem.GetComponent<ParticipantsListItem>().Setup(clientUUID, clientName, isMasterClient);
        }
    }
    private void HandleUpdateStartCountdown(int countdown)
    {
        if (countdown > 0)
        {
            txtStartCountdown.text = countdown.ToString();
            txtStartCountdown.gameObject.SetActive(true);
            // 카운트 다운 중간엔 시작 버튼이 눌릴 수 없음
            btnStartGame.interactable = false;
        }
        else
        {
            txtStartCountdown.gameObject.SetActive(false);
        }
    }
    private void HandleEnableStart()
    {
        btnStartGame.interactable = true;
    }
    private void HandleStartGame(Texture2D spriteSheet)
    {
        currentQuizImageSpritesheet = spriteSheet;
    }
    private void HandleUpdateQuizCountdown(int countdown)
    {
        if (countdown > 0)
        {
            txtQuizCountdown.text = "시작 " + countdown.ToString();
            txtQuizCountdown.gameObject.SetActive(true);
        }
        else
        {
            txtQuizCountdown.gameObject.SetActive(false);
        }
    }

    private void HandleShowQuizImage(int currentQuizIndex)
    {
        ShowQuizImage(currentQuizIndex);

        // 퀴즈 카테고리 표시 로직
        for (int i = 0; i < gameAnswerCategoryNames.Count; i++)
        {
            // 정답 초기화
            txtGameAnswers[i].text = "_";
            lblGameAnswers[i].text = gameAnswerCategoryNames[i];
        }
    }
    private void HandleStartQuiz(List<string> categoryNames)
    {
        // 이미지가 보여질 때 같이 카테고리가 보일 수 있게 미리 값 저장
        gameAnswerCategoryNames = categoryNames;
    }

    private void HandleUpdateQuizTimer(int remainingTime)
    {
        txtQuizTimer.text = "남은 시간: " + remainingTime.ToString() + "초";
        Color currColor = txtQuizTimer.color;

        if (remainingTime <= 5)
        {
            if(currColor != Color.red)
                txtQuizTimer.color = Color.red;
        }
        else
        {
            if(currColor != Color.black)
                txtQuizTimer.color = Color.black;
        }
    }

    private void HandleRevealQuizAnswer(List<string> keywords)
    {
        // 정답 표시 로직
        for (int i = 0; i < keywords.Count; i++)
        {
            txtGameAnswers[i].text = keywords[i];
        }
    }

    private void HandleOccurQuizAnswer(string clientNickname, string category_name, string keyword)
    {
        // TODO: 정답 UI 오브젝트에 반영
        //     + 지금 category_name을 받는 걸 category_id로 바꿔야할 수도 있음
        Debug.Log($"{clientNickname}님이 카테고리 {category_name}의 정답 \"{keyword}\"을 맞췄습니다.");
    }

    private void HandleEndGame(Dictionary<string, int> scores)
    {
        txtFinalScore.text = "";
        grpScoreBoard.SetActive(true);

        // 점수판에 점수 표시
        foreach (var score in scores)
        {
            txtFinalScore.text += $"{score.Key}: {score.Value}점\n";
        }
    }
    


    public void OnClickedCreateRoom()
    {
        grpCreateRoom.SetActive(true);
    }
    public void OnClickedExitCreateRoom()
    {
        grpCreateRoom.SetActive(false);
    }
    public void OnClickedConfirmCreateRoom()
    {
        int maxParticipants = GameManager.Instance.SelectionManager.GetSelectedMaxParticipants();
        string roomName = inputRoomName.text;

        if (!string.IsNullOrEmpty(roomName) && maxParticipants != -1)
        {
            GameManager.Instance.WebSocketClient.CreateRoom(roomName, maxParticipants);
            InitGroupCreateRoom();
        }
        else
        {
            Debug.LogWarning("방 제목이 입력되지 않았거나 최대 인원 수가 선택되지 않았습니다.");
        }
    }
    public void OnClickedJoinRoom()
    {
        string selectedRoomUUID = GameManager.Instance.SelectionManager.GetSelectedRoomUUID();
        if (!string.IsNullOrEmpty(selectedRoomUUID))
        {
            GameManager.Instance.WebSocketClient.JoinRoom(selectedRoomUUID);
        }
        else
        {
            // TODO : 버튼이 비활성화된 상태여야함
            Debug.LogWarning("No room selected.");
        }
    }
    public void OnClickedLeftRoom()
    {
        GameManager.Instance.WebSocketClient.LeftRoom();
    }
    public void OnClickedStartGame()
    {
        GameManager.Instance.WebSocketClient.StartGame();
    }

    void InitGroupCreateRoom()
    {
        grpCreateRoom.SetActive(false);
        inputRoomName.text = "";
        InitLayMaxParticipants();
    }
    void InitLayMaxParticipants()
    {
        GameManager.RemoveAllChildren(layMaxParticipants);

        const int maximumParticipants = 6; // TODO: 서버에서 관리하거나 헤더 파일에서 관리하도록
        for (int i = 2; i <= maximumParticipants; i++)
        {
            GameObject maxParticipantsItem = Instantiate(prefabMaxParticipantsItem, layMaxParticipants.transform);

            maxParticipantsItem.GetComponent<MaxParitipantsItem>().Setup(i);
        }
    }

    // # In Game
    private void ShowQuizImage(int quizIndex)
    {
        int x = (quizIndex % 3) * 512;
        int y = (quizIndex / 3) * 512;
        Rect spriteRect = new Rect(x, y, 512, 512);

        // 현재 quizIndex에 걸맞는 Sprite 생성
        Sprite quizSprite = Sprite.Create(currentQuizImageSpritesheet, spriteRect, new Vector2(0.5f, 0.5f));

        // 스프라이트 설정
        imgQuiz.sprite = quizSprite;

        // 이미지 크기 조정
        float imageSize = Mathf.Min(imgQuiz.rectTransform.rect.width, imgQuiz.rectTransform.rect.height);
        imgQuiz.rectTransform.sizeDelta = new Vector2(imageSize, imageSize);

        // 이미지 크기 조정 모드 설정
        imgQuiz.preserveAspect = true;
    }

    private void OnClickedOkScoreBoard()
    {
        grpScoreBoard.SetActive(false);
        GameManager.Instance.SetState(GameState.InRoom);
    }
}
