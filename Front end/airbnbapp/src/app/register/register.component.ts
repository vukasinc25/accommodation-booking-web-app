import { Component } from '@angular/core';
import { FormGroup, FormBuilder, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../service/auth.service';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
})
export class RegisterComponent {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private authService: AuthService
  ) {
    this.form = this.fb.group({
      username: [null, Validators.required],
      email: [null, Validators.required],
      firstName: [null, Validators.required],
      lastName: [null, Validators.required],
      country: [null, Validators.required],
      city: [null, Validators.required],
      streetName: [null, Validators.required],
      streetNumber: [null, Validators.required],
      password: [null, Validators.required],
      userRole: [null, Validators.required],
    });
  }

  submit() {
    this.authService.register(this.form.value).subscribe({
      next: (data) => {
        console.log('Register success');
        this.router.navigate(['login']);
      },
      error: (err) => {
        console.log(err);
      },
    });
  }
}
