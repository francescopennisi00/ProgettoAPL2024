﻿using CommunityToolkit.Mvvm.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using System.Windows.Input;
using WeatherClient.Exceptions;
using WeatherClient.Views;

namespace WeatherClient.ViewModels;

internal class RuleViewModel : ObservableObject, IQueryAttributable
{
    private Models.Rule _rule;

    public bool IsEnableDeleteCommand { get; private set; }

    private bool _isEditableLocation;
    private bool _isNotEditableLocation;

    public bool IsEditableLocation 
    { 
        get
        {
            return _isEditableLocation;
        }
        private set
        {
            if (value != _isEditableLocation)
            {
                _isEditableLocation = value;
                OnPropertyChanged();
            }
        }
    }

    public bool IsNotEditableLocation 
    {
        get => _isNotEditableLocation;
        private set
        {
            if (value != _isNotEditableLocation)
            {
                _isNotEditableLocation = value;
                OnPropertyChanged();
            }
        }
    }

    public string? Id
    {
        get
        {
            return _rule.Id;
        }
        private set
        {
            _rule.Id = value;
        }
    }

    public string? LocationName
    {
        get
        {
            if (_rule.Location == null)
            {
                return null;
            }
            return _rule.Location[0];
        }
        set
        {
            if (_rule.Location == null)
            {
                _rule.Location = new List<string>(new string[5]);
            }
            _rule.Location[0] = value;
            OnPropertyChanged();
        }
    }

    public string? LocationCountry
    {
        get
        {
            if (_rule.Location == null)
            {
                return null;
            }
            return _rule.Location[3];
        }
        set
        {
            if (_rule.Location == null)
            {
                _rule.Location = new List<string>(new string[5]);
            }
            _rule.Location[3] = value;
            OnPropertyChanged();
        }
    }

    public string? LocationState
    {
        get
        {
            if (_rule.Location == null)
            {
                return null;
            }
            return _rule.Location[4];
        }
        set
        {
            if (_rule.Location == null)
            {
                _rule.Location = new List<string>(new string[5]);
            }
            _rule.Location[4] = value;
            OnPropertyChanged();
        }
    }

    public string? LocationLatitude
    {
        get
        {
            if (_rule.Location == null)
            {
                return null;
            }
            return _rule.Location[1];
        }
        set
        {
            if (_rule.Location == null)
            {
                _rule.Location = new List<string>(new string[5]);
            }
            _rule.Location[1] = value;
            OnPropertyChanged();
        }
    }

    public string? LocationLongitude
    {
        get
        {
            if (_rule.Location == null)
            {
                return null;
            }
            return _rule.Location[2];
        }
        set
        {
            if (_rule.Location == null)
            {
                _rule.Location = new List<string>(new string[5]);
            }
            _rule.Location[2] = value;
            OnPropertyChanged();
        }
    }

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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
            if (value != null && value != "")
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
        IsEditableLocation = false;
        IsNotEditableLocation = true;
        SaveCommand = new AsyncRelayCommand(Save);
        DeleteCommand = new AsyncRelayCommand(Delete);
    }

    public RuleViewModel(Models.Rule rule)
    {
        _rule = rule;
        IsEditableLocation = false;
        IsNotEditableLocation = true;
        SaveCommand = new AsyncRelayCommand(Save);
        DeleteCommand = new AsyncRelayCommand(Delete);
    }

    private async Task Save()
    {
        try
        {
            if(TriggerPeriod.Length == 0 || LocationLatitude.Length == 0 || LocationLongitude.Length == 0 || LocationName.Length == 0)
            {
                var title = "Alert!";
                var message = "You have to insert all required fields";
                await Application.Current.MainPage.DisplayAlert(title, message, "OK");
                return;
            }
            // In case of success, id contains either id of the _rule created into server or the string returned by server
            // we are interested only in id (so we use id string variable only id Id == null)
            string id = await _rule.Save();
            // if Id property is null, then saved _rule is new and we have to assign to Id property its id
            if (Id == String.Empty)
            {
                Id = id;
            }
        }
        catch (BadRequestException exc)
        {
            var title = "Warning!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(AllRulesPage));
        }
        catch (TokenNotValidException exc)
        {
            var title = "Login Required!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync("//LoginRoute");
        }
        catch (ServerException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(AllRulesPage));
        }
        await Shell.Current.GoToAsync($"..?saved={_rule.Id}");
    }

    private async Task Delete()
    {
        try
        {
            _rule.Delete();
        }
        catch (BadRequestException exc)
        {
            var title = "Warning!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(AllRulesPage));
        }
        catch (TokenNotValidException exc)
        {
            var title = "Login Required!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync("//LoginRoute");
        }
        catch (ServerException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(AllRulesPage));
        }
        await Shell.Current.GoToAsync($"..?deleted={_rule.Id}");
    }

    async void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("load"))
        {
            IsEnableDeleteCommand = true;
            try
            {
                _rule = await Models.Rule.Load(query["load"].ToString());
                RefreshProperties();
            } 
            catch (TokenNotValidException exc)
            {
                var title = "Login Required!";
                var message = exc.Errormessage;
                await Application.Current.MainPage.DisplayAlert(title, message, "OK");
                await Shell.Current.GoToAsync("//LoginRoute");
            }
            catch (ServerException exc)
            {
                var title = "Error!";
                var message = exc.Errormessage;
                await Application.Current.MainPage.DisplayAlert(title, message, "OK");
                await Shell.Current.GoToAsync(nameof(AllRulesPage));
            }
            catch (Exception)
            {
                var title = "Error!";
                var message = "Error in loading _rule.";
                await Application.Current.MainPage.DisplayAlert(title, message, "OK");
                await Shell.Current.GoToAsync(nameof(AllRulesPage));
            }
        }
        if (query.ContainsKey("add"))
        {
            IsEditableLocation = true;
            IsNotEditableLocation = false;
            IsEnableDeleteCommand = false;
        }
    }

    public async void Reload()
    {
        try
        {
            _rule = await Models.Rule.Load(_rule.Id);
        }
        catch (TokenNotValidException exc)
        {
            var title = "Login Required!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync("//LoginRoute");
        }
        catch (ServerException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(AllRulesPage));
        }
        catch (Exception)
        {
            var title = "Error!";
            var message = "Error in loading _rule.";
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
            await Shell.Current.GoToAsync(nameof(AllRulesPage));
        }
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