using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class MaxParitipantsItem : MonoBehaviour
{
    public int MaxParticipants {  get; private set; }
    
    Button _button = null;

    public void Setup(int maxParticipants)
    {
        MaxParticipants = maxParticipants;
        gameObject.name = maxParticipants.ToString();
        GetComponentInChildren<TextMeshProUGUI>().text = maxParticipants.ToString();
        _button = GetComponent<Button>();
        _button.onClick.AddListener(OnClick);
    }
    void OnClick()
    {
        GameManager.Instance.SelectionManager.SelectMaxParticipantsItem(this);
    }
    public void Highlight(bool highlight)
    {
        _button.interactable = !highlight; // 하이라이트된 버튼은 클릭 불가 -> 자동으로 disable 컬러로 전환됨
    }
}
