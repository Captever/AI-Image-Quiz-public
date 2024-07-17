using UnityEngine;

public class SingletonObj<T> : MonoBehaviour where T : MonoBehaviour
{
    private static T _instance = null;
    private static readonly object _lock = new object();
    private static bool _isShuttingDown = false;

    public static T Instance
    {
        get
        {
            if (_isShuttingDown)
            {
                Debug.LogWarning("[Singleton] Instance '" + typeof(T) + "' already destoryed. Returning null.");
                return null;
            }

            lock (_lock)
            {
                if(_instance == null)
                {
                    _instance = (T)FindObjectOfType(typeof(T));

                    if(_instance == null)
                    {
                        var singletonObject = new GameObject();
                        _instance = singletonObject.AddComponent<T>();
                        singletonObject.name = typeof(T).ToString() + " (Singleton)";

                        DontDestroyOnLoad(singletonObject);
                    }
                    else
                    {
                        DontDestroyOnLoad(_instance.gameObject);
                    }
                }
                
                return _instance;
            }
        }
    }

    private void OnApplicationQuit()
    {
        _isShuttingDown = true;
    }

    private void OnDestroy()
    {
        _isShuttingDown = true;
    }
}