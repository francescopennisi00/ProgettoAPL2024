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

namespace WeatherClient.Models;

internal class Rule
{
    [JsonProperty("rules")]
    public WeatherParameters? Rules { get; set; }

    [JsonProperty("trigger_period")]
    public string? TriggerPeriod { get; set; }

    [JsonProperty("location")]
    public List<string>? Location { get; set; }

    [JsonProperty("id")]
    public string? Id { get; set; }


    public Rule()
    {
        Rules = new WeatherParameters();
        TriggerPeriod = string.Empty;
        List<string> Location = new();
        Id = string.Empty;
    }

    public Rule(string id, string locName, string lat, string lon, string cCode, string sCode, string triggerP, string maxT,
        string minT, string maxH, string minH, string maxP, string minP, string maxWS, string minWS, string direction, string rain,
        string snow, string maxC, string minC)
    {
        Rules = new WeatherParameters(maxT, minT, maxH, minH, maxP, minP, maxWS, minWS, direction, rain, snow, maxC, minC);
        TriggerPeriod = triggerP;
        Location = new List<string> { locName, lat, lon, cCode, sCode };
        Id = id;
    }

    private static string GetToken()
    {
        // Get the folder where the tokes is stored.
        string appDataPath = FileSystem.AppDataDirectory + @"\JWT_token.txt";
        string token = File.ReadAllText(appDataPath);
        return token;
    }

    private async Task<int> DoRequest(string url)
    {
        string token = GetToken();
        HttpClient httpC = new HttpClient();
        httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer ", token);
        string jsonData = Utilities.JsonUtility.SerializeJSON(this);
        var content = new StringContent(jsonData, Encoding.UTF8, "application/json");
        HttpResponseMessage response = await httpC.PostAsync(url, content);
        return (int)response.StatusCode;
    }

    public async Task<int> Save()
    {
        return await DoRequest(Utilities.Constants.urlUpdate);
    }

    public async Task<int> Delete()
    {
        return await DoRequest(Utilities.Constants.urlDelete);
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
        string url = Constants.urlShow;
        string token = GetToken();
        HttpClient httpC = new HttpClient();
        httpC.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer ", token);
        HttpResponseMessage response = await httpC.GetAsync(url);
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

    public static IEnumerable<Rule> LoadAll()
    {
        Task <IEnumerable<Rule>> task = LoadAllAsync();
        IEnumerable<Rule> rules = task.Result;
        return rules;
    }

}