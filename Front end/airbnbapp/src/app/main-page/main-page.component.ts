import { Component, OnInit } from '@angular/core';
import { Accommodation } from '../model/accommodation';
import { AccommodationService } from '../service/accommodation.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-main-page',
  templateUrl: './main-page.component.html',
  styleUrls: ['./main-page.component.css'],
})
export class MainPageComponent implements OnInit {
  constructor(
    private router: Router,
    private accommodationService: AccommodationService
  ) {}

  accommodations: Accommodation[] = [];

  ngOnInit(): void {
    this.accommodationService.getAll().subscribe({
      next: (data) => {
        this.accommodations = data as Accommodation[];
        console.log(this.accommodations);
      },
      error: (err) => {
        console.log(err);
      },
    });
  }
}
