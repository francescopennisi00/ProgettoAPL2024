using System.ComponentModel;
using System.Windows.Input;
using WeatherClient.Exceptions;

namespace WeatherClient.ViewModels;

internal class SignupViewModel : INotifyPropertyChanged
{
    private bool isLoading;
    private Models.User _user;

    public string UserName
    {
        get => _user.UserName;
        set => _user.UserName = value;
    }
    public string Password
    {
        get => _user.Password;
        set => _user.Password = value;
    }


    //proprietà per indicare che la finestra sta caricando
    public bool IsLoading
    {
        get => isLoading;
        set
        {
            isLoading = value;
            NotifyPropertyChanged(nameof(IsLoading));
            NotifyPropertyChanged(nameof(IsNotLoading));
        }
    }

    public bool IsNotLoading
    {
        get => !isLoading;
    }


    public string ConfirmPassword { get; set; }
    public ICommand Signup { get; set; }

    public SignupViewModel(string username) : this()
    {
        UserName = username;
    }

    public SignupViewModel()
    {
        _user = new Models.User();
        Signup = new Command(SignupClicked);
    }

    public event PropertyChangedEventHandler PropertyChanged;

    //metodo per notificare gli observer che osservano l'evento PropertyChanged
    private void NotifyPropertyChanged(string propertyName)
    {
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
    }

    private async void SignupClicked()
    {
        //controlli di correttezza dei dati
        if (string.IsNullOrEmpty(UserName) || string.IsNullOrEmpty(Password))
        {
            await App.Current.MainPage.DisplayAlert("Attenzione!", "Riempire i campi username e password", "Ok");
            return;
        }
        if (ConfirmPassword != Password)
        {
            await App.Current.MainPage.DisplayAlert("Attenzione!", "Il campo password e conferma password devonon coincidere", "Ok");
            return;
        }

        IsLoading = true;
        try
        {
            if (await _user.SignUp())
            {
                if (await _user.Login())
                {
                    await Shell.Current.GoToAsync($"..?registered={UserName}");
                    await Shell.Current.GoToAsync("//AllRulesRoute");
                }
                else
                    await App.Current.MainPage.DisplayAlert("Error", "Username or password wrong. Retry!", "Ok");
            }
            else
                await App.Current.MainPage.DisplayAlert("Error", "Email already in use. Try to sign in!", "Ok");
        }
        catch (EmailAlreadyInUseException exc)
        {
            var title = "Error!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
        catch (ServerException exc)
        {
            var title = "Warning!";
            var message = exc.Errormessage;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
        IsLoading = false;
    }


}
