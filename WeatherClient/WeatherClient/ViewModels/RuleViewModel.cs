using CommunityToolkit.Mvvm.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using System.Windows.Input;
using WeatherClient.Exceptions;

namespace WeatherClient.ViewModels;

internal class RuleViewModel : ObservableObject, IQueryAttributable
{
    private Models.Rule _rule;

    public string? TriggerPeriod => _rule.TriggerPeriod;
    public List<string>? Location => _rule.Location;
    public string? Id => _rule.Id;

    public bool IsMaxTempVisible { get; set; }
    public string? MaxTemp
    {
        get
        { 
            if (_rule.Rules.MaxTemp == "null")
            {
                IsMaxTempVisible = false;
                OnPropertyChanged(nameof(IsMaxTempVisible));
            } else
            {
                IsMaxTempVisible = true;
                OnPropertyChanged(nameof(IsMaxTempVisible));
            }
            return _rule.Rules.MaxTemp;
        }
        set
        {
            if (_rule.Rules.MaxTemp != value)
            {
                _rule.Rules.MaxTemp = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMinTempVisible { get; set; }
    public string? MinTemp
    {
        get
        {
            if (_rule.Rules.MinTemp == "null")
            {
                IsMinTempVisible = false;
                OnPropertyChanged(nameof(IsMinTempVisible));
            }
            else
            {
                IsMinTempVisible = true;
                OnPropertyChanged(nameof(IsMinTempVisible));
            }
            return _rule.Rules.MinTemp;
        }
        set
        {
            if (_rule.Rules.MinTemp != value)
            {
                _rule.Rules.MinTemp = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMaxHumVisible { get; set; }
    public string? MaxHumidity
    {
        get
        {
            if (_rule.Rules.MaxHumidity == "null")
            {
                IsMaxHumVisible = false;
                OnPropertyChanged(nameof(IsMaxHumVisible));
            }
            else
            {
                IsMaxHumVisible = true;
                OnPropertyChanged(nameof(IsMaxHumVisible));
            }
            return _rule.Rules.MaxHumidity;
        }
        set
        {
            if (_rule.Rules.MaxHumidity != value)
            {
                _rule.Rules.MaxHumidity = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMinHumVisible { get; set; }
    public string? MinHumidity
    {
        get
        {
            if (_rule.Rules.MinHumidity == "null")
            {
                IsMinHumVisible = false;
                OnPropertyChanged(nameof(IsMinHumVisible));
            }
            else
            {
                IsMinHumVisible = true;
                OnPropertyChanged(nameof(IsMinHumVisible));
            }
            return _rule.Rules.MinHumidity;
        }
        set
        {
            if (_rule.Rules.MinHumidity != value)
            {
                _rule.Rules.MinHumidity = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMaxPressVisible { get; set; }
    public string? MaxPressure
    {
        get
        {
            if (_rule.Rules.MaxPressure == "null")
            {
                IsMaxPressVisible = false;
                OnPropertyChanged(nameof(IsMaxPressVisible));
            }
            else
            {
                IsMaxPressVisible = true;
                OnPropertyChanged(nameof(IsMaxPressVisible));
            }
            return _rule.Rules.MaxPressure;
        }
        set
        {
            if (_rule.Rules.MaxPressure != value)
            {
                _rule.Rules.MaxPressure = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMinPressVisible { get; set; }
    public string? MinPressure
    {
        get
        {
            if (_rule.Rules.MinPressure == "null")
            {
                IsMinPressVisible = false;
                OnPropertyChanged(nameof(IsMinPressVisible));
            }
            else
            {
                IsMinPressVisible = true;
                OnPropertyChanged(nameof(IsMinPressVisible));
            }
            return _rule.Rules.MinPressure;
        }
        set
        {
            if (_rule.Rules.MinPressure != value)
            {
                _rule.Rules.MinPressure = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMaxWindSpeedVisible { get; set; }
    public string? MaxWindSpeed
    {
        get
        {
            if (_rule.Rules.MaxWindSpeed == "null")
            {
                IsMaxWindSpeedVisible = false;
                OnPropertyChanged(nameof(IsMaxWindSpeedVisible));
            }
            else
            {
                IsMaxWindSpeedVisible = true;
                OnPropertyChanged(nameof(IsMaxWindSpeedVisible));
            }
            return _rule.Rules.MaxWindSpeed;
        }
        set
        {
            if (_rule.Rules.MaxWindSpeed != value)
            {
                _rule.Rules.MaxWindSpeed = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMinWindSpeedVisible { get; set; }
    public string? MinWindSpeed
    {
        get
        {
            if (_rule.Rules.MinWindSpeed == "null")
            {
                IsMinWindSpeedVisible = false;
                OnPropertyChanged(nameof(IsMinWindSpeedVisible));
            }
            else
            {
                IsMinWindSpeedVisible = true;
                OnPropertyChanged(nameof(IsMinWindSpeedVisible));
            }
            return _rule.Rules.MinWindSpeed;
        }
        set
        {
            if (_rule.Rules.MinWindSpeed != value)
            {
                _rule.Rules.MinWindSpeed = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsWindDirectionVisible { get; set; }
    public string? WindDirection
    {
        get
        {
            if (_rule.Rules.WindDirection == "null")
            {
                IsWindDirectionVisible = false;
                OnPropertyChanged(nameof(IsWindDirectionVisible));
            }
            else
            {
                IsWindDirectionVisible = true;
                OnPropertyChanged(nameof(IsWindDirectionVisible));
            }
            return _rule.Rules.WindDirection;
        }
        set
        {
            if (_rule.Rules.WindDirection != value)
            {
                _rule.Rules.WindDirection = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsRainVisible { get; set; }
    public bool? Rain
    {
        get
        {
            if (_rule.Rules.Rain == "null")
            {
                IsRainVisible = false;
                OnPropertyChanged(nameof(IsRainVisible));
                return false;
            }
            else
            {
                IsRainVisible = true;
                OnPropertyChanged(nameof(IsRainVisible));
                return true;
            }
        }
        set
        {
            if (_rule.Rules.Rain == "null")
            {
                if (value == true)
                {
                    _rule.Rules.Rain = "rain";
                    OnPropertyChanged();
                }
            }
            else
            {
                if (value == false)
                {
                    _rule.Rules.Rain = "null";
                    OnPropertyChanged();
                }
            }
        }
    }

    public bool IsSnowVisible { get; set; }
    public bool? Snow
    {
        get
        {
            if (_rule.Rules.Snow == "null")
            {
                IsSnowVisible = false;
                OnPropertyChanged(nameof(IsSnowVisible));
                return false;
            }
            else
            {
                IsSnowVisible = true;
                OnPropertyChanged(nameof(IsSnowVisible));
                return true;
            }
        }
        set
        {
            if (_rule.Rules.Snow == "null")
            {
                if (value == true)
                {
                    _rule.Rules.Snow = "snow";
                    OnPropertyChanged();
                }
            }
            else
            {
                if (value == false)
                {
                    _rule.Rules.Snow = "null";
                    OnPropertyChanged();
                }
            }
        }
    }

    public bool IsMaxCloudVisible { get; set; }
    public string? MaxCloud
    {
        get
        {
            if (_rule.Rules.MaxCloud == "null")
            {
                IsMaxCloudVisible = false;
                OnPropertyChanged(nameof(IsMaxCloudVisible));
            }
            else
            {
                IsMaxCloudVisible = true;
                OnPropertyChanged(nameof(IsMaxCloudVisible));
            }
            return _rule.Rules.MaxCloud;
        }
        set
        {
            if (_rule.Rules.MaxCloud != value)
            {
                _rule.Rules.MaxCloud = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsMinCloudVisible { get; set; }
    public string? MinCloud
    {
        get
        {
            if (_rule.Rules.MinCloud == "null")
            {
                IsMinCloudVisible = false;
                OnPropertyChanged(nameof(IsMinCloudVisible));
            }
            else
            {
                IsMinCloudVisible = true;
                OnPropertyChanged(nameof(IsMinCloudVisible));
            }
            return _rule.Rules.MinCloud;
        }
        set
        {
            if (_rule.Rules.MinCloud != value)
            {
                _rule.Rules.MinCloud = value;
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
        } catch (TokenNotValidException exc)
        {
            /*creare pannello x andare alla login*/
        } catch (ServerException exc)
        {
            /* creare pannello per tornare alla home rules*/
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
            /*creare pannello x andare alla login*/
        }
        catch (ServerException exc)
        {
            /* creare pannello per tornare alla home rules*/
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