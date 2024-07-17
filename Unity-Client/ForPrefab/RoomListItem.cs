using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class RoomListItem : MonoBehaviour
{
    public string RoomUUID { get; private set; }
    public Client MasterClient {  get; private set; }

    Button _button = null;

    [SerializeField] private TextMeshProUGUI txtRoomTitle;
    [SerializeField] private TextMeshProUGUI txtMasterClientName;

    public void Setup(string roomUUID, string roomTitle, Client masterClient)
    {
        RoomUUID = roomUUID;
        MasterClient = masterClient;
        gameObject.name = roomUUID;

        txtRoomTitle.text = roomTitle;
        txtMasterClientName.text = masterClient.ClientNickname;

        _button = GetComponent<Button>();
        _button.onClick.AddListener(OnClick);
    }
    void OnClick()
    {
        GameManager.Instance.SelectionManager.SelectRoomListItem(this);
    }
    public void Highlight(bool highlight)
    {
        _button.interactable = !highlight; // 하이라이트된 버튼은 클릭 불가 -> 자동으로 disable 컬러로 전환됨
    }
    // 게임 중일 시 색상을 따로 적용할 때 사용
    public void SetInGameColor()
    {
        GetComponent<Image>().color = Color.red;
    }
}
