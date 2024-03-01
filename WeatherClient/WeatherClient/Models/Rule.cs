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
namespace WeatherClient.Models;

internal class Rule
{
    [JsonProperty("rules")]
    public WeatherParameters Rules { get; private set; }

    [JsonProperty("trigger_period")]
    public string TriggerPeriod { get; set; }

    [JsonProperty("location")]
    public string[] Location { get; set; }

    [JsonProperty("id")]
    public string Id { get; set; }


    public Rule()
    {
    }

    /*public void Save() =>
        File.WriteAllText(System.IO.Path.Combine(FileSystem.AppDataDirectory, Filename), Text);
    */
    public void Delete(string Id) {
        string url = "http://weather.com:9090/update_rules/delete_user_constraints_by_location";
        string token = GetToken();
        using (HttpClient client = new HttpClient())
        {
            client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer ", token);

            HttpResponseMessage response = await client.GetAsync(url);

            if (response.IsSuccessStatusCode)
            {
                string responseContent = await response.Content.ReadAsStringAsync();
                // Supponendo che il contenuto della risposta sia una lista di regole in formato JSON
                // E che tu abbia una classe Rule con un metodo statico Parse per convertire il JSON in un oggetto Rule
                return Utilities.JsonUtility.DeserializeJSON<IEnumerable<Rule>>(responseContent);
            }
            else if (response.StatusCode == HttpStatusCode.Unauthorized)
            {
                await Shell.Current.GoToAsync(nameof(Views.LoginPage));
                return Enumerable.Empty<Rule>(); // Restituisci una collezione vuota in caso di Unauthorized
            }
            else
            {
                string responseContent = await response.Content.ReadAsStringAsync();
                // Gestione degli errori, ad esempio log o lancio di un'eccezione
                throw new Exception($"Failed to load rules. Status code: {response.StatusCode},{responseContent}");
            }
        }
    }
    public static Rule Load(string LocationName)
    {
        filename = System.IO.Path.Combine(FileSystem.AppDataDirectory, filename);

        if (!File.Exists(filename))
            throw new FileNotFoundException("Unable to find file on local storage.", filename);

        return
            new()
            {
                Rules = 
            };
    }

    public static string GetToken()
    {
        // Get the folder where the tokes is stored.
        string appDataPath = FileSystem.AppDataDirectory + @"\JWT_token.txt";
        string token = File.ReadAllText(appDataPath);
        return token;
    }

    public static async Task<IEnumerable<Rule>> LoadAllAsync()
    {
        string url = "http://weather.com:9090/show_rules";
        string token = GetToken();
        using (HttpClient client = new HttpClient())
        {
            client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer ", token);

            HttpResponseMessage response = await client.GetAsync(url);

            if (response.IsSuccessStatusCode)
            {
                string responseContent = await response.Content.ReadAsStringAsync();
                // Supponendo che il contenuto della risposta sia una lista di regole in formato JSON
                // E che tu abbia una classe Rule con un metodo statico Parse per convertire il JSON in un oggetto Rule
                return Utilities.JsonUtility.DeserializeJSON<IEnumerable<Rule>>(responseContent);
            }
            else if (response.StatusCode == HttpStatusCode.Unauthorized)
            {
                await Shell.Current.GoToAsync(nameof(Views.LoginPage));
                return Enumerable.Empty<Rule>(); // Restituisci una collezione vuota in caso di Unauthorized
            }
            else
            {   
                string responseContent = await response.Content.ReadAsStringAsync();
                // Gestione degli errori, ad esempio log o lancio di un'eccezione
                throw new Exception($"Failed to load rules. Status code: {response.StatusCode},{responseContent}");
            }
        }
    }
    public static IEnumerable<Rule> LoadAll()
    {
        Task <IEnumerable<Rule>> task = LoadAllAsync();
        IEnumerable<Models.Rule> rules = task.Result;
        return rules;
    }

}
