import { Component, OnInit } from '@angular/core';
import { AuthService } from '../service/auth.service';
import { Router } from '@angular/router';
import { ProfServiceService } from '../service/prof.service.service';

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css'],
})
export class ProfileComponent implements OnInit {
  user = {
    username: '',
    email: '',
    firstName: '',
    lastName: '',
    city: '',
    country: '',
    streetName: '',
    streetNumber: '',
  };
  constructor(
    private authService: AuthService,
    private profService: ProfServiceService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.profService.getUserInfo().subscribe({
      next: (data) => {
        this.user.username = data.username;
        this.user.email = data.email;
        this.user.firstName = data.firstname;
        this.user.lastName = data.lastname;
        this.user.city = data.location.city;
        this.user.country = data.location.country;
        this.user.streetName = data.location.streetName;
        this.user.streetNumber = data.location.streetNumber;
      },
      error: (err) => {
        alert(err.error.message);
      },
    });
  }

  submitForm() {
    this.profService.updateUserInfo(this.user).subscribe({
      next: (data) => {
        alert('User succesfully updated');
        this.router.navigate(['']);
      },
      error: (err) => {
        alert(err.error.message);
      },
    });
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['']);
  }
}
