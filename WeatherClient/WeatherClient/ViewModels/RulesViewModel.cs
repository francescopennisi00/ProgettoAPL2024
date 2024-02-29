using CommunityToolkit.Mvvm.Input;
using WeatherClient.Models;
using System.Collections.ObjectModel;
using System.Windows.Input;

namespace WeatherClient.ViewModels;

internal class RulesViewModel : IQueryAttributable
{
    public ObservableCollection<ViewModels.RuleViewModel> AllRules { get; }
    public ICommand NewCommand { get; }
    public ICommand SelectNoteCommand { get; }
    public RulesViewModel()
    {
        AllRules = new ObservableCollection<ViewModels.RuleViewModel>(Models.Rule.LoadAllAsync().Select(n => new RuleViewModel(n)));
        NewCommand = new AsyncRelayCommand(NewNoteAsync);
        SelectNoteCommand = new AsyncRelayCommand<ViewModels.RuleViewModel>(SelectNoteAsync);
    }

    private async Task NewNoteAsync()
    {
        await Shell.Current.GoToAsync(nameof(Views.RulesPage));
    }

    private async Task SelectNoteAsync(ViewModels.RuleViewModel note)
    {
        if (note != null)
            await Shell.Current.GoToAsync($"{nameof(Views.RulesPage)}?load={note.Identifier}");
    }

    void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("deleted"))
        {
            string noteId = query["deleted"].ToString();
            RuleViewModel matchedNote = AllRules.Where((n) => n.Identifier == noteId).FirstOrDefault();

            // If note exists, delete it
            if (matchedNote != null)
                AllRules.Remove(matchedNote);
        }
        else if (query.ContainsKey("saved"))
        {
            string noteId = query["saved"].ToString();
            RuleViewModel matchedNote = AllRules.Where((n) => n.Identifier == noteId).FirstOrDefault();

            // If note is found, update it
            if (matchedNote != null)
            {
                matchedNote.Reload();
                AllRules.Move(AllRules.IndexOf(matchedNote), 0);
            }

            // If note isn't found, it's new; add it.
            else
                AllRules.Insert(0, new RuleViewModel(Models.Rule.Load(noteId)));
        }
    }
}