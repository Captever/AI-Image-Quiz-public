using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.UI;

public class InitSceneManager : MonoBehaviour
{
    [SerializeField] private TMP_InputField inputNickname;
    [SerializeField] private Button btnConfirmNickname;

    // Start is called before the first frame update
    void Awake()
    {
        Screen.SetResolution(450, 975, false); // 원하는 해상도로 변경

        inputNickname.text = "";

        btnConfirmNickname.onClick.AddListener(OnConfirmNickname);
    }

    void OnConfirmNickname()
    {
        // TODO: 닉네임 중복 확인 및 네트워크 연결 확인
        GameManager.Instance.SetNickname(inputNickname.text);

        GameManager.Instance.WebSocketClient.ConnectToMaster(); // 마스터 서버에 접속

        GameManager.Instance.LoadNextScene();
    }
}
