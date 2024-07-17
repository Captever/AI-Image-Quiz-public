using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class ChatManager : MonoBehaviour
{
    [SerializeField] private TextMeshProUGUI txtChatLogs;
    [SerializeField] private TMP_InputField inputChatMessage;
    [SerializeField] private Button btnSendChat;


    void Awake()
    {
        GameManager.Instance.SetScriptChatManager(this);
        btnSendChat.onClick.AddListener(OnClickedSendChat);
        inputChatMessage.onEndEdit.AddListener(OnEndedInputChat);
    }
    private void OnEnable()
    {
        txtChatLogs.text = "";
        inputChatMessage.text = "";
    }
    void OnEndedInputChat(string text)
    {
        // 실질적인 입력이 있을 때 Enter 입력이 발생하면 요청
        if (!string.IsNullOrEmpty(inputChatMessage.text) && (
                Input.GetKeyDown(KeyCode.Return) || 
                Input.GetKeyDown(KeyCode.KeypadEnter)))
        {
            SendChatMessage();
        }
    }
    void OnClickedSendChat()
    {
        // 실질적인 입력이 있을 때만 요청
        if (!string.IsNullOrEmpty(inputChatMessage.text))
        {
            SendChatMessage();
        }
    }
    void SendChatMessage()
    {
        GameManager.Instance.WebSocketClient.SendChatMessage(inputChatMessage.text);
        inputChatMessage.text = "";
        inputChatMessage.ActivateInputField();
    }

    public void AddNewChatMessage(string newMessage)
    {
        // TODO: 나중에 채팅 메시지 별 prefab으로 나눠서
        txtChatLogs.text = newMessage + "\n" + txtChatLogs.text;
    }
    public void AddNewSystemMessage(string newMessage)
    {
        // TODO: 나중에 시스템 메시지 prefab으로 적용
        txtChatLogs.text = newMessage + "\n" + txtChatLogs.text;
    }
}
