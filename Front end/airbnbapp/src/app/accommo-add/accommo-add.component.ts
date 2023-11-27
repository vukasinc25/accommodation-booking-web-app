import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AccommodationService } from 'src/app/service/accommodation.service';

@Component({
  selector: 'app-accommo-add',
  templateUrl: './accommo-add.component.html',
  styleUrls: ['./accommo-add.component.css'],
})
export class AccommoAddComponent {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private accommodationService: AccommodationService
  ) {
    this.form = this.fb.group({
      name: [null, Validators.required],
      location: [null, Validators.required],
      amenities: [null, Validators.required],
      minGuests: [null, Validators.required],
      maxGuests: [null, Validators.required],
      price: [null, Validators.required],
    });
  }

  submit() {
    this.accommodationService.insert(this.form.value).subscribe({
      next: (data) => {
        console.log('create success');
      },
      error: (err) => {
        console.log(err);
      },
    });
  }
}
