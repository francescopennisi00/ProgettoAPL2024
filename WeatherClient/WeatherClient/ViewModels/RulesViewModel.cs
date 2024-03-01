using CommunityToolkit.Mvvm.Input;
using WeatherClient.Models;
using System.Collections.ObjectModel;
using System.Windows.Input;

namespace WeatherClient.ViewModels;

internal class RulesViewModel : IQueryAttributable
{
    public ObservableCollection<RuleViewModel> AllRules { get; }
    public ICommand NewCommand { get; }
    public ICommand SelectNoteCommand { get; }
    public RulesViewModel()
    {
        AllRules = new ObservableCollection<RuleViewModel>(Rule.LoadAll().Select(n => new RuleViewModel(n)));
        NewCommand = new AsyncRelayCommand(NewNoteAsync);
        SelectNoteCommand = new AsyncRelayCommand<RuleViewModel>(SelectNoteAsync);
    }

    private async Task NewNoteAsync()
    {
        await Shell.Current.GoToAsync(nameof(Views.RulesPage));
    }

    private async Task SelectNoteAsync(RuleViewModel rule)
    {
        if (rule != null)
            await Shell.Current.GoToAsync($"{nameof(Views.RulesPage)}?load={rule.Id}");
    }

    void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("deleted"))
        {
            string LocationID = query["deleted"].ToString();
            RuleViewModel matcheRule = AllRules.Where((n) => n.Location[0] == LocationID).FirstOrDefault();

            // If rule exists, delete it
            if (matcheRule != null)
                AllRules.Remove(matcheRule);
        }
        else if (query.ContainsKey("saved"))
        {
            string LocationID = query["saved"].ToString();
            RuleViewModel matchedRule = AllRules.Where((n) => n.Location[0] == LocationID).FirstOrDefault();

            // If rule is found, update it
            if (matchedRule != null)
            {
                matchedRule.Reload();
                AllRules.Move(AllRules.IndexOf(matchedRule), 0);
            }

            // If rule isn't found, it's new; add it.
            else
                AllRules.Insert(0, new RuleViewModel(Models.Rule.Load(LocationID)));
        }
    }
}