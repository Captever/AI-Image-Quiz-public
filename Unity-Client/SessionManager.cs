using System.Collections;
using TMPro;
using UnityEngine;
using UnityEngine.Networking;
using UnityEngine.UI;
using System.Security.Cryptography;
using System.Text;

public class SessionManager : MonoBehaviour
{
    // # Error Message
    [SerializeField] private GameObject grpAlert;
    [SerializeField] private Image icoAlert;
    [SerializeField] private Sprite imgInfo;
    [SerializeField] private Sprite imgError;
    [SerializeField] private TextMeshProUGUI txtAlert;
    [SerializeField] private Button btnOkAlert;

    // # Sign up
    [SerializeField] private Button btnSignup;
    [SerializeField] private GameObject grpSignup;
    [SerializeField] private Button btnExitSignup;
    [SerializeField] private TMP_InputField inputSignupEmail;
    [SerializeField] private TMP_InputField inputSignupUsername;
    [SerializeField] private TMP_InputField inputSignupPassword;
    [SerializeField] private Button btnConfirmSignup;

    // # Sign in
    [SerializeField] private GameObject grpSignin;
    [SerializeField] private TMP_InputField inputSigninUsername;
    [SerializeField] private TMP_InputField inputSigninPassword;
    [SerializeField] private Button btnConfirmSignin;


    private void Awake()
    {
        // # init alert
        grpAlert.SetActive(false);
        txtAlert.text = "";
        btnOkAlert.onClick.AddListener(OnClickedOkAlert);

        // # init sign up
        inputSignupEmail.text = "";
        inputSignupUsername.text = "";
        inputSignupPassword.text = "";
        btnSignup.onClick.AddListener(OnClickedSignup);
        btnExitSignup.onClick.AddListener(OnClickedExitSignup);
        btnConfirmSignup.onClick.AddListener(OnClickedConfirmSignup);


        // # init sign in
        inputSigninUsername.text = "";
        inputSigninPassword.text = "";
        btnConfirmSignin.onClick.AddListener(OnClickedConfirmSignin);

        SetUISignin(true);
    }

    // # OnClick
    void OnClickedOkAlert()
    {
        grpAlert.SetActive(false);
    }
    void OnClickedSignup()
    {
        SetUISignin(false);
    }
    void OnClickedExitSignup()
    {
        SetUISignin(true);
    }
    public void OnClickedConfirmSignup()
    {
        string email = inputSignupEmail.text;
        string username = inputSignupUsername.text;
        string password = inputSignupPassword.text;
        StartCoroutine(SignupCoroutine(email, username, password));
    }
    public void OnClickedConfirmSignin()
    {
        string username = inputSigninUsername.text;
        string password = inputSigninPassword.text;
        StartCoroutine(SigninCoroutine(username, password));
    }

    // # UI
    void SetUISignin(bool toSignin)
    {
        grpSignin.SetActive(toSignin);
        grpSignup.SetActive(!toSignin);
    }
    void ShowAlert(string message, bool isError = true)
    {
        if (isError)
        {
            icoAlert.sprite = imgError;
            icoAlert.color = Color.red;
            txtAlert.color = Color.red;
        }
        else
        {
            icoAlert.sprite = imgInfo;
            icoAlert.color = Color.blue;
            txtAlert.color = Color.black;
        }

        txtAlert.text = message;
        grpAlert.SetActive(true);
    }

    private string HashPassword(string password)
    {
        using (SHA256 sha256 = SHA256.Create())
        {
            byte[] hashedBytes = sha256.ComputeHash(Encoding.UTF8.GetBytes(password));
            return BitConverter.ToString(hashedBytes).Replace("-", "").ToLower();
        }
    }

    // TODO: SignupCoroutine, SigninCoroutine 통합하기
    private IEnumerator SigninCoroutine(string username, string password)
    {
        string signinUrl = "http://" + Config.ServerUrl + "/signin";
        Debug.Log("로그인 시도 중: " + signinUrl);
        string hashedPassword = HashPassword(password);
        SigninRequest signinRequest = new SigninRequest { username = username, password = password };
        string jsonData = JsonUtility.ToJson(signinRequest);

        UnityWebRequest request = new UnityWebRequest(signinUrl, "POST");
        byte[] bodyRaw = System.Text.Encoding.UTF8.GetBytes(jsonData);
        request.uploadHandler = new UploadHandlerRaw(bodyRaw);
        request.downloadHandler = new DownloadHandlerBuffer();
        request.SetRequestHeader("Content-Type", "application/json");

        yield return request.SendWebRequest();

        if (request.result == UnityWebRequest.Result.Success)
        {
            var response = JsonUtility.FromJson<SigninResponse>(request.downloadHandler.text);
            PlayerPrefs.SetString("SessionUUID", response.SessionUUID);
            PlayerPrefs.SetString("ClientUUID", response.ClientUUID);
            PlayerPrefs.Save();

            ShowAlert("로그인이 성공했습니다.", false);
            Debug.Log($"로그인 성공: {response.SessionUUID}");
            grpSignin.SetActive(false);
            GameManager.Instance.WebSocketClient.ConnectToServer();
        }
        else
        {
            ShowAlert("로그인에 실패했습니다.");
            Debug.LogError("Signin failed: " + request.error);
        }
    }
    private IEnumerator SignupCoroutine(string email, string username, string password)
    {
        string signupUrl = "http://" + Config.ServerUrl + "/signup";
        Debug.Log("회원가입 시도 중: " + signupUrl);
        string hashedPassword = HashPassword(password);
        SignupRequest signupRequest = new SignupRequest { email = email, username = username, password = password };
        string jsonData = JsonUtility.ToJson(signupRequest);

        UnityWebRequest request = new UnityWebRequest(signupUrl, "POST");
        byte[] bodyRaw = System.Text.Encoding.UTF8.GetBytes(jsonData);
        request.uploadHandler = new UploadHandlerRaw(bodyRaw);
        request.downloadHandler = new DownloadHandlerBuffer();
        request.SetRequestHeader("Content-Type", "application/json");

        yield return request.SendWebRequest();

        if (request.result == UnityWebRequest.Result.Success)
        {
            ShowAlert("회원 가입이 성공했습니다.", false);
            SetUISignin(true);
        }
        else
        {
            ShowAlert("회원 가입을 실패했습니다.");
            Debug.LogError("Signup failed: " + request.error);
        }
    }

    [System.Serializable]
    private class SigninRequest
    {
        public string username;
        public string password;
    }
    [System.Serializable]
    private class SignupRequest
    {
        public string email;
        public string username;
        public string password;
    }
    [System.Serializable]
    private class SigninResponse
    {
        public string SessionUUID;
        public string ClientUUID;
    }
}