import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AccommodationService } from '../service/accommodation.service';
import { Accommodation } from '../model/accommodation';

@Component({
  selector: 'app-accommo-info',
  templateUrl: './accommo-info.component.html',
  styleUrls: ['./accommo-info.component.css'],
})
export class AccommoInfoComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    private accommodationService: AccommodationService
  ) {}

  id: number = 0;
  accommodation: Accommodation = {};
  isDataEmpty = false;

  ngOnInit(): void {
    this.route.params.subscribe((params) => {
      this.id = params['id'];
    });

    this.accommodationService.getById(this.id).subscribe({
      next: (data) => {
        this.accommodation = data;
        console.log(data);
      },
      error: (err) => {
        console.log(err);
        this.isDataEmpty = true;
      },
    });
  }
}
