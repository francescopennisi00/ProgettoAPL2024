using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Net.Http;
using System.Threading.Tasks;
using System.Net;
using System.Collections.ObjectModel;
using Newtonsoft.Json;
using WeatherClient.Utilities;
using WeatherClient.Exceptions;
using Windows.Media.Protection.PlayReady;

namespace WeatherClient.Models;

internal class Rule
{
    /*[JsonProperty("rules")]
    public WeatherParameters? Rules { get; set; }*/
    [JsonProperty("max_temp")]
    public string? MaxTemp { get; set; }

    [JsonProperty("min_temp")]
    public string? MinTemp { get; set; }

    [JsonProperty("max_humidity")]
    public string? MaxHumidity { get; set; }

    [JsonProperty("min_humidity")]
    public string? MinHumidity { get; set; }

    [JsonProperty("max_pressure")]
    public string? MaxPressure { get; set; }

    [JsonProperty("min_pressure")]
    public string? MinPressure { get; set; }

    [JsonProperty("max_wind_speed")]
    public string? MaxWindSpeed { get; set; }

    [JsonProperty("min_wind_speed")]
    public string? MinWindSpeed { get; set; }

    [JsonProperty("wind_direction")]
    public string? WindDirection { get; set; }

    [JsonProperty("rain")]
    public string? Rain { get; set; }

    [JsonProperty("snow")]
    public string? Snow { get; set; }

    [JsonProperty("max_cloud")]
    public string? MaxCloud { get; set; }

    [JsonProperty("min_cloud")]
    public string? MinCloud { get; set; }

    [JsonProperty("trigger_period")]
    public string? TriggerPeriod { get; set; }

    [JsonProperty("location")]
    public List<string>? Location { get; set; }

    [JsonProperty("id")]
    public string? Id { get; set; }


    public Rule()
    {
        /*Rules = new WeatherParameters();*/
        MaxTemp = "null";
        MinTemp = "null";
        MaxHumidity = "null";
        MinHumidity = "null";
        MaxPressure = "null";
        MinPressure = "null";
        MaxWindSpeed = "null";
        MinWindSpeed = "null";
        WindDirection = "null";
        Rain = "null";
        Snow = "null";
        MaxCloud = "null";
        MinCloud = "null";
        TriggerPeriod = string.Empty;
        List<string> Location = new();
        Id = string.Empty;
    }

    public Rule(string id, string locName, string lat, string lon, string cCode, string sCode, string triggerP, string maxT,
        string minT, string maxH, string minH, string maxP, string minP, string maxWS, string minWS, string direction, string rain,
        string snow, string maxC, string minC)
    {
        //Rules = new WeatherParameters(maxT, minT, maxH, minH, maxP, minP, maxWS, minWS, direction, rain, snow, maxC, minC);
        MaxTemp = maxT;
        MinTemp = minT;
        MaxHumidity = maxH;
        MinHumidity = minH;
        MaxPressure = maxP;
        MinPressure = minP;
        MaxWindSpeed = maxWS;
        MinWindSpeed = minWS;
        WindDirection = direction;
        Rain = rain;
        Snow = snow;
        MaxCloud = maxC;
        MinCloud = minC;
        TriggerPeriod = triggerP;
        Location = new List<string> { locName, lat, lon, cCode, sCode };
        Id = id;
    }

    private static string GetToken()
    {
        // Get the folder where the tokes is stored.
        // string appDataPath = FileSystem.AppDataDirectory + @"\JWT_token.txt";
        string appDataPath = @"C:\Users\Utente\Desktop\token.txt";
        string token = File.ReadAllText(appDataPath);
        return token;
    }

    private async Task DoRequest(string url)
    {
        string token = GetToken();
        HttpClient httpC = new HttpClient();
        httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", token);
        string jsonData = Utilities.JsonUtility.SerializeJSON(this);
        HttpRequestMessage request = new HttpRequestMessage
        {
            Method = HttpMethod.Post,
            RequestUri = new Uri(url),
            Content = new StringContent(jsonData, Encoding.UTF8, "application/json")
        };
        var prova = request.Content;
        HttpResponseMessage response = httpC.Send(request);
        if ((int)response.StatusCode == 401)
        {
            throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
        }
        else if ((int)response.StatusCode != 200)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
    }

    public async void Save()
    {
        await DoRequest(Utilities.Constants.urlUpdate);
    }

    public async void Delete()
    {
        await DoRequest(Utilities.Constants.urlDelete);
    }

    public static Rule Load(string id)
    {
        IEnumerable<Rule> rules = LoadAll();
        foreach (Rule rule in rules)
        {
            if (rule.Id == id)
            {
                return new()
                {
                    Rules = rule.Rules,
                    Id = rule.Id,
                    TriggerPeriod = rule.TriggerPeriod,
                    Location = rule.Location
                };
            }
        }
        throw new Exception("Error! Rule not found!");
    }
 

    public static async Task<IEnumerable<Rule>> LoadAllAsync()
    {
        try
        {
            string url = Constants.urlShow;
            string token = GetToken();
            HttpClient httpC = new HttpClient();
            httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", token);
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = HttpMethod.Get,
                RequestUri = new Uri(url),
                Content = null
            };
            HttpResponseMessage response = httpC.Send(request);

            if (response.IsSuccessStatusCode)
            {
                string responseContent = await response.Content.ReadAsStringAsync();
                return JsonUtility.DeserializeJSON<IEnumerable<Rule>>(responseContent);
            }
            else if (response.StatusCode == HttpStatusCode.Unauthorized)
            {
                throw new TokenNotValidException("JWT Token provided is not valid. Login required.");
            }
            else
            {
                throw new ServerException("Failed to load rules due an internal server error.");
            }
        }
        catch (HttpRequestException)
        {
            throw new ServerException("Failed to load rules due an internal server error.");
        }
    }

    public static IEnumerable<Rule> LoadAll()
    {
        Task <IEnumerable<Rule>> task = LoadAllAsync();
        IEnumerable<Rule> rules = task.Result;
        if( rules == null)
        {
            return new List<Rule>();
        }
        return rules;
    }
    /*
    private async Task<HttpResponseMessage> DoRequest(HttpMethod httpreq, string url, string json)
    {
        try
        {
            HttpRequestMessage request = new HttpRequestMessage
            {
                Method = httpVerb,
                RequestUri = new Uri(url),
                Content = null
            };
            if (!string.IsNullOrEmpty(UserService.Instance.Token))
                richiesta.Headers.Add("Authorization", UserService.Instance.Token); //aggiungo l'eventuale token, se disponibile
            if (!string.IsNullOrEmpty(json))
                richiesta.Content = new StringContent(json, Encoding.UTF8, "application/json"); //aggiungo l'eventuale contenuto

            HttpResponseMessage risposta = await _client.SendAsync(richiesta);

            return risposta;
        }
        catch (Exception e)
        {
            return new HttpResponseMessage(HttpStatusCode.BadRequest);
        }
    }
    */
}