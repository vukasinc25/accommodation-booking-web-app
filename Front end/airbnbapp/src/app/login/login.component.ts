import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../service/auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
})
export class LoginComponent {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private authService: AuthService
  ) {
    this.form = this.fb.group({
      username: [null, Validators.required],
      password: [null, Validators.required],
    });
  }

  submit() {
    this.authService.login(this.form.value).subscribe({
      next: (data) => {
        console.log('login success');
        console.log(data);
        localStorage.setItem('jwt', data.access_token);
        this.router.navigate(['']);
      },
      error: (err) => {
        console.log(err);
      },
    });
  }
}
