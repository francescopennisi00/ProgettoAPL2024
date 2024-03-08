using CommunityToolkit.Mvvm.Input;
using WeatherClient.Models;
using System.Collections.ObjectModel;
using System.Windows.Input;
using WeatherClient.Exceptions;

namespace WeatherClient.ViewModels;

internal class RulesViewModel : IQueryAttributable
{
    private Rule rule;

    public ObservableCollection<RuleViewModel> AllRules { get; private set; }
    public ICommand NewCommand { get; }
    public ICommand SelectNoteCommand { get; }

    public async void LoadAllRules()
    {
        try
        {
            var results = await Rule.LoadAll();
            AllRules = new ObservableCollection<RuleViewModel>(results.Select(n => new RuleViewModel(n)));
        }
        catch (TokenNotValidException exc)
        {
            var title = "Warning!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync("//LoginRoute");
        }
        catch (ServerException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
    }

    public RulesViewModel()
    {
        LoadAllRules();
        NewCommand = new AsyncRelayCommand(NewRuleAsync);
        SelectNoteCommand = new AsyncRelayCommand<RuleViewModel>(SelectRuleAsync);
    }

    public RulesViewModel(Rule rule)
    {
        this.rule = rule;
    }

    private async Task NewRuleAsync()
    {
        await Shell.Current.GoToAsync($"{nameof(Views.RulesPage)}?add={true}");
    }

    private async Task SelectRuleAsync(RuleViewModel rule)
    {
        if (rule != null)
            await Shell.Current.GoToAsync($"{nameof(Views.RulesPage)}?load={rule.Id}");
    }

    void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("deleted"))
        {
            string id = query["deleted"].ToString();
            RuleViewModel matcheRule = AllRules.Where((n) => n.Id == id).FirstOrDefault();

            // If rule exists, delete it
            if (matcheRule != null)
                AllRules.Remove(matcheRule);
        }
        else if (query.ContainsKey("saved"))
        {
            string id = query["saved"].ToString();
            RuleViewModel matchedRule = AllRules.Where((n) => n.Id == id).FirstOrDefault();

            // If rule is found, update it
            if (matchedRule != null)
            {
                matchedRule.Reload();
                AllRules.Move(AllRules.IndexOf(matchedRule), 0);
            }

            // If rule isn't found, it's new; add it.
            else
            {
                try
                {
                    AllRules.Insert(0, new RuleViewModel(Rule.Load(id)));
                }
                catch (ServerException ex)
                {
                    var title = "Error!";
                    Application.Current.MainPage.DisplayAlert(title, ex.Message, "OK");   
                }
                catch (TokenNotValidException ex)
                {
                    var title = "Error!";
                    Application.Current.MainPage.DisplayAlert(title, ex.Message, "OK");
                    Shell.Current.GoToAsync("//LoginRoute");

                }
                catch (Exception ex)
                {
                    var title = "Warning!";
                    Application.Current.MainPage.DisplayAlert(title, ex.Message, "OK");
                }
            }
        }
    }
}