using UnityEngine;

public class SelectionManager : MonoBehaviour
{
    // == RoomList ==
    private RoomListItem selectedRoomListItem;

    public void SelectRoomListItem(RoomListItem roomListItem)
    {
        // 이전 선택된 방을 해제
        if (selectedRoomListItem != null)
        {
            selectedRoomListItem.Highlight(false);
        }

        // 새로운 방 선택
        selectedRoomListItem = roomListItem;
        selectedRoomListItem.Highlight(true);
    }

    public string GetSelectedRoomUUID()
    {
        return selectedRoomListItem.RoomUUID;
    }


    // == MaxParticipants ==
    private MaxParitipantsItem selectedMaxParticipantsItem;

    public void SelectMaxParticipantsItem(MaxParitipantsItem maxParticipantsItem)
    {
        // 이전 선택된 방을 해제
        if (selectedMaxParticipantsItem != null)
        {
            selectedMaxParticipantsItem.Highlight(false);
        }

        // 새로운 방 선택
        selectedMaxParticipantsItem = maxParticipantsItem;
        selectedMaxParticipantsItem.Highlight(true);
    }

    public int GetSelectedMaxParticipants()
    {
        return selectedMaxParticipantsItem.MaxParticipants;
    }
}
