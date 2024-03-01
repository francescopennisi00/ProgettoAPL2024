using Newtonsoft.Json;

namespace WeatherClient.Utilities;

public static class JsonUtility
{
    public static string SerializeJSON(object obj) => JsonConvert.SerializeObject(obj);

    public static T? DeserializeJSON<T>(string str) => JsonConvert.DeserializeObject<T>(str);
}