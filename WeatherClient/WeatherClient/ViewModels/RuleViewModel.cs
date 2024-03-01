using CommunityToolkit.Mvvm.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using System.Windows.Input;
using WeatherClient.Utilities;

namespace WeatherClient.ViewModels;

internal class RuleViewModel : ObservableObject, IQueryAttributable
{
    private Models.Rule _rule;
    public WeatherParameters Rules => _rule.Rules;

    public string TriggerPeriod => _rule.TriggerPeriod;
    public string[] Location => _rule.Location;

    public ICommand SaveCommand { get; private set; }
    public ICommand DeleteCommand { get; private set; }
    public RuleViewModel()
    {
        _rule = new Models.Rule();
        SaveCommand = new AsyncRelayCommand(Save);
        DeleteCommand = new AsyncRelayCommand(Delete);
    }

    public RuleViewModel(Models.Rule rule)
    {
        _rule = rule;
        SaveCommand = new AsyncRelayCommand(Save);
        DeleteCommand = new AsyncRelayCommand(Delete);
    }
    public string MaxTemp
    {
        get => _rule.Rules.MaxTemp;
        set
        {
            if (_rule.Rules.MaxTemp != value)
            {
                _rule.Rules.MaxTemp = value;
                OnPropertyChanged();
            }
        }
    }
    private async Task Save()
    {
        _rule.Date = DateTime.Now;
        _rule.Save();
        await Shell.Current.GoToAsync($"..?saved={_rule.Filename}");
    }

    private async Task Delete()
    {
        _rule.Delete();
        await Shell.Current.GoToAsync($"..?deleted={_rule.Id}");
    }

    void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("load"))
        {
            _rule = Models.Rule.Load(query["load"].ToString());
            RefreshProperties();
        }
    }

    public void Reload()
    {
        _rule = Models.Rule.Load(_rule.Filename);
        RefreshProperties();
    }

    private void RefreshProperties()
    {
        OnPropertyChanged(nameof(Text));
        OnPropertyChanged(nameof(Date));
    }


}