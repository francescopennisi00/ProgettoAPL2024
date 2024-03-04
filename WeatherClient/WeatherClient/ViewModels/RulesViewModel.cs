using CommunityToolkit.Mvvm.Input;
using WeatherClient.Models;
using System.Collections.ObjectModel;
using System.Windows.Input;
using WeatherClient.Exceptions;

namespace WeatherClient.ViewModels;

internal class RulesViewModel : IQueryAttributable
{
    public ObservableCollection<RuleViewModel> AllRules { get; }
    public ICommand NewCommand { get; }
    public ICommand SelectNoteCommand { get; }
    public RulesViewModel()
    {
        try
        {
            AllRules = new ObservableCollection<RuleViewModel>(Rule.LoadAll().Result.Select(n => new RuleViewModel(n)));
        }
        catch (TokenNotValidException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
        catch (ServerException exc)
        {
            var title = "Warning!";
            var message = exc.Errormessage;
            Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
        catch (AggregateException)
        {
            var title = "Warning!";
            var message = "Ofancul!";
            Application.Current.MainPage.DisplayAlert(title, message, "OK");
            Shell.Current.GoToAsync("//LoginRoute");
        }
        NewCommand = new AsyncRelayCommand(NewRuleAsync);
        SelectNoteCommand = new AsyncRelayCommand<RuleViewModel>(SelectRuleAsync);
    }

    private async Task NewRuleAsync()
    {
        await Shell.Current.GoToAsync(nameof(Views.RulesPage));
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
                AllRules.Insert(0, new RuleViewModel(Models.Rule.Load(id)));
        }
    }
}