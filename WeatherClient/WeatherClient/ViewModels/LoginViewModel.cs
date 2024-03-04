using CommunityToolkit.Mvvm.ComponentModel;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Input;
using WeatherClient.Exceptions;
using WeatherClient.Models;
using WeatherClient.Views;
namespace WeatherClient.ViewModels;

internal class LoginViewModel : ObservableObject, IQueryAttributable
{
    private Models.User _user;
    public bool IsVisibleLogin { get; private set; }
    public bool IsVisibleLogout { get; private set; }
    public string UserName
    {
        get => _user.UserName;
        set{
            _user.UserName = value;
            OnPropertyChanged(nameof(UserName));
        }
    }
    public string Password
    {
        get => _user.Password;
        set => _user.Password = value;
    }

    public ICommand Login { get; set; }
    public ICommand Signup { get; set; }
    public ICommand Logout { get; set; }
    public ICommand DeleteAccount { get; set; }
    public LoginViewModel()
    {
        _user = new Models.User();
        IsVisibleLogin = true;
        IsVisibleLogout = false;
        Login = new Command(LoginClicked);

        Signup = new Command(async () =>
        {
            SignupPage page = new SignupPage();
            if (!string.IsNullOrEmpty(UserName)) //se l'utente ha già inserito l'username nel login lo facilito nel signup inserendolo in automatico nella nuova pagina
                page.BindingContext = new SignupViewModel(UserName);
            await App.Current.MainPage.Navigation.PushAsync(page);
        });
        Logout = new Command(LogoutClicked);
        DeleteAccount = new Command(DeleteAccountClicked);
    }
    void IQueryAttributable.ApplyQueryAttributes(IDictionary<string, object> query)
    {
        if (query.ContainsKey("registered"))
        {
            IsVisibleLogin = false;
            IsVisibleLogout = true;
            OnPropertyChanged(nameof(IsVisibleLogin));
            OnPropertyChanged(nameof(IsVisibleLogout));
            var username = query["registered"].ToString();
            if (!string.IsNullOrEmpty(username))
            {
                UserName = username;
            }
        }
        
    }
    private async void LogoutClicked()
    {
        try
        {
            _user.Logout();
            IsVisibleLogin = true;
            IsVisibleLogout = false;
            OnPropertyChanged(nameof(IsVisibleLogin));
            OnPropertyChanged(nameof(IsVisibleLogout));
            _user.Password = String.Empty;
            _user.UserName = String.Empty;
            OnPropertyChanged(nameof(Password));
            OnPropertyChanged(nameof(UserName));
        }
        catch (Exception exc)
        {
            var title = "Error!";
            var message = exc.Message;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }
    }

    private async void LoginClicked()
    {
        try
        {
            if (await _user.Login())
            {
                IsVisibleLogin = false;
                IsVisibleLogout = true;
                OnPropertyChanged(nameof(IsVisibleLogin));
                OnPropertyChanged(nameof(IsVisibleLogout));
                //se il login ha successo setto le credenziali e visualizzo la tabbedpage
                await Shell.Current.GoToAsync("//AllRulesRoute");
            }
            else
            {
                await App.Current.MainPage.DisplayAlert("Error", "Username or password wrong. Retry!", "Ok");
            }
        }
        catch (UsernamePswWrongException exc)
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

    }
    private async void DeleteAccountClicked()
    {
        try
        {
            _user.DeleteAccount();
            IsVisibleLogin = true;
            IsVisibleLogout = false;
            OnPropertyChanged(nameof(IsVisibleLogin));
            OnPropertyChanged(nameof(IsVisibleLogout));
            _user.Password = String.Empty;
            _user.UserName = String.Empty;
            OnPropertyChanged(nameof(Password));
            OnPropertyChanged(nameof(UserName));
        }
        catch (Exception exc)
        {
            var title = "Error!";
            var message = exc.Message;
            await Application.Current.MainPage.DisplayAlert(title, message, "OK");
        }

    }
}
