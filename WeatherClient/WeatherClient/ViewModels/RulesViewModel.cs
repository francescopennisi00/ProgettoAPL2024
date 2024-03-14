using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using System.Collections.ObjectModel;
using System.Windows.Input;
using WeatherClient.Exceptions;
using WeatherClient.Models;
using WeatherClient.Views;

namespace WeatherClient.ViewModels;

internal class RulesViewModel : ObservableObject, IQueryAttributable
{
    private Rule _rule;
    public ObservableCollection<RuleViewModel> AllRules { get; private set; }

    private bool _activeAddCommand = false;
    public bool ActiveAddCommand
    {
        get
        {
            return _activeAddCommand;
        }
        set
        {
            if (value != _activeAddCommand)
            {
                _activeAddCommand = value;
                OnPropertyChanged();
            }
        }
    }

    public ICommand NewCommand { get; }
    public ICommand SelectNoteCommand { get; }

    public async void LoadAllRules()
    {
        try
        {
            var results = await Rule.LoadAll();
            foreach (var result in results)
            {
                AllRules.Add(new RuleViewModel(result));
            }
            // if all rules was successfull we enable add command
            ActiveAddCommand = true;
        }
        catch (TokenNotValidException exc)
        {
            var title = "Error!";
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
        AllRules = new ObservableCollection<RuleViewModel>();
        LoadAllRules();
        NewCommand = new AsyncRelayCommand(NewRuleAsync);
        SelectNoteCommand = new AsyncRelayCommand<RuleViewModel>(SelectRuleAsync);
    }

    public RulesViewModel(Rule rule)
    {
        this._rule = rule;
    }

    private async Task NewRuleAsync()
    {
        await Shell.Current.GoToAsync($"{nameof(RulesPage)}?add={true}");
    }

    private async Task SelectRuleAsync(RuleViewModel rule)
    {
        if (rule != null)
            await Shell.Current.GoToAsync($"{nameof(RulesPage)}?load={rule.Id}");
    }

    async void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("deleted"))
        {
            string id = query["deleted"].ToString();
            RuleViewModel matcheRule = AllRules.Where((n) => n.Id == id).FirstOrDefault();

            // If _rule exists, delete it
            if (matcheRule != null)
                AllRules.Remove(matcheRule);
        }
        else if (query.ContainsKey("saved"))
        {
            string id = query["saved"].ToString();
            RuleViewModel matchedRule = AllRules.Where((n) => n.Id == id).FirstOrDefault();

            // If _rule is found, update it
            if (matchedRule != null)
            {
                matchedRule.Reload();
                AllRules.Move(AllRules.IndexOf(matchedRule), 0);
            }

            // If _rule isn't found, it's new; add it.
            else
            {
                try
                {
                    Rule ruleToAdd = await Rule.Load(id);
                    AllRules.Insert(0, new RuleViewModel(ruleToAdd));
                }
                catch (ServerException ex)
                {
                    var title = "Error!";
                    Application.Current.MainPage.DisplayAlert(title, ex.Message, "OK");
                }
                catch (TokenNotValidException ex)
                {
                    var title = "Login Required!";
                    Application.Current.MainPage.DisplayAlert(title, ex.Message, "OK");
                    await Shell.Current.GoToAsync("//LoginRoute");

                }
                catch (Exception ex)
                {
                    var title = "Warning!";
                    Application.Current.MainPage.DisplayAlert(title, ex.Message, "OK");
                }
            }
        }
        else if (query.ContainsKey("logout"))
        {
            AllRules.Clear();
            OnPropertyChanged(nameof(AllRules));
            ActiveAddCommand = false;
        }
        else if (query.ContainsKey("login"))
        {
            if (AllRules.Count == 0)
            {
                LoadAllRules();
                OnPropertyChanged(nameof(AllRules));
            }
        }
    }
}