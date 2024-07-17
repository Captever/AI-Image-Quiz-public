using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class ParticipantsListItem : MonoBehaviour
{
    public string ClientUUID { get; private set; }
    public string ClientName { get; private set; }

    Button _button = null;

    [SerializeField] private TextMeshProUGUI txtClientName;
    [SerializeField] private GameObject objIsMasterClient;

    public void Setup(string clientUUID, string clientName, bool isMasterClient)
    {
        ClientUUID = clientUUID;
        ClientName = clientName;
        gameObject.name = clientUUID;

        txtClientName.text = clientName;
        objIsMasterClient.SetActive(isMasterClient);

        _button = GetComponent<Button>();
        _button.onClick.AddListener(OnClick);
    }
    void OnClick()
    {
        // TODO: 참가자 선택지에 맞게 변경
        Debug.Log("'" + ClientName + "'을 선택하셨습니다.");
    }
    public void Highlight(bool highlight)
    {
        _button.interactable = !highlight; // 하이라이트된 버튼은 클릭 불가 -> 자동으로 disable 컬러로 전환됨
    }
}
