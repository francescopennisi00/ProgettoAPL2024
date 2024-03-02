using CommunityToolkit.Mvvm.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using System.Windows.Input;
using WeatherClient.Exceptions;

namespace WeatherClient.ViewModels;

internal class RuleViewModel : ObservableObject, IQueryAttributable
{
    private Models.Rule _rule;

    public List<string>? Location => _rule.Location;
    public string? Id => _rule.Id;

    public string? TriggerPeriod
    {
        get => _rule.TriggerPeriod;
        set
        {
            _rule.TriggerPeriod = value;
            OnPropertyChanged();
            
        }
    }

    public string? MaxTemp
    {
        get
        {
            if (_rule.Rules.MaxTemp == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MaxTemp;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MaxTemp = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MaxTemp = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MinTemp
    {
        get
        {
            if (_rule.Rules.MinTemp == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MinTemp;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MinTemp = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MinTemp = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MaxHumidity
    {
        get
        {
            if (_rule.Rules.MaxHumidity == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MaxHumidity;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MaxHumidity = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MaxHumidity = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MinHumidity
    {
        get
        {
            if (_rule.Rules.MinHumidity == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MinHumidity;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MinHumidity = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MinHumidity = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MaxPressure
    {
        get
        {
            if (_rule.Rules.MaxPressure == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MaxPressure;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MaxPressure = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MaxPressure = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MinPressure
    {
        get
        {
            if (_rule.Rules.MinPressure == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MinPressure;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MinPressure = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MinPressure = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MaxWindSpeed
    {
        get
        {
            if (_rule.Rules.MaxWindSpeed == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MaxWindSpeed;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MaxWindSpeed = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MaxWindSpeed = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MinWindSpeed
    {
        get
        {
            if (_rule.Rules.MinWindSpeed == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MinWindSpeed;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MinWindSpeed = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MinWindSpeed = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? WindDirection
    {
        get
        {
            if (_rule.Rules.WindDirection == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.WindDirection;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.WindDirection = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.WindDirection = "null";
                OnPropertyChanged();
            }
        }
    }

    public bool? Rain
    {
        get
        {
            if (_rule.Rules.Rain == "null")
            {
                return false;
            }
            else
            {
                return true;
            }
        }
        set
        {
            if (value == true)
            {
                _rule.Rules.Rain = "rain";
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.Rain = "null";
                OnPropertyChanged();
            }
        }
    }

    public bool? Snow
    {
        get
        {
            if (_rule.Rules.Snow == "null")
            {
                return false;
            }
            else
            {
                return true;
            }
        }
        set
        {

            if (value == true)
            {
                _rule.Rules.Snow = "snow";
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.Snow = "null";
                OnPropertyChanged();
            }
            
        }
    }

    public string? MaxCloud
    {
        get
        {
            if( _rule.Rules.MaxCloud == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MaxCloud;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MaxCloud = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MaxCloud = "null";
                OnPropertyChanged();
            }
        }
    }

    public string? MinCloud
    {
        get
        {
            if (_rule.Rules.MinCloud == "null")
            {
                return null;
            }
            else
            {
                return _rule.Rules.MinCloud;
            }
        }
        set
        {
            if (value != null)
            {
                _rule.Rules.MinCloud = value;
                OnPropertyChanged();
            }
            else
            {
                _rule.Rules.MinCloud = "null";
                OnPropertyChanged();
            }
        }
    }

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

    private async Task Save()
    {
        try
        {
            _rule.Save();
        }
        catch (TokenNotValidException exc)
        {
            var title = "Login Required!";
            var message = "Your session is expired. You will be redirect to login page.";
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(Views.LoginPage));
        }
        catch (ServerException exc)
        {
            var title = "Warning!";
            var message = "An error occurred in server. It was not possible to save your rule.";
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(Views.AllRulesPage));
        }
        await Shell.Current.GoToAsync($"..?saved={_rule.Id}");
    }

    private async Task Delete()
    {
        try
        {
            _rule.Delete();
        }
        catch (TokenNotValidException exc)
        {
            var title = "Login Required!";
            var message = "Your session is expired. You will be redirect to login page.";
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(Views.LoginPage));
        }
        catch (ServerException exc)
        {
            var title = "Warning!";
            var message = "An error occurred in server. It was not possible to delete your rule.";
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(Views.AllRulesPage));
        }
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
        _rule = Models.Rule.Load(_rule.Id);
        RefreshProperties();
    }

    private void RefreshProperties()
    {
        var properties = GetType().GetProperties();

        foreach (var property in properties)
        {
            OnPropertyChanged(property.Name);
        }
    }


}