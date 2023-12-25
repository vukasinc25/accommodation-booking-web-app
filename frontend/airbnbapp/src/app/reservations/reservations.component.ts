import { ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { ReservationService } from '../service/reservation.service';
import { AccommodationService } from '../service/accommodation.service';

@Component({
  selector: 'app-reservations',
  templateUrl: './reservations.component.html',
  styleUrls: ['./reservations.component.css'],
})
export class ReservationsComponent implements OnInit {
  userReservations: any[] = [];

  constructor(
    private router: Router,
    private accommodationService: AccommodationService,
    private reservationService: ReservationService
  ) {}

  ngOnInit(): void {
    this.reservationService.getAllReservationsByUserId().subscribe({
      next: (data) => {
        this.userReservations = data;
      },
      error: (err) => {
        alert('Cant get reservation for user');
        console.log(err.error.error);
        this.router.navigate(['']);
      },
    });
  }

  undoReservation(reservation: any) {
    this.reservationService.cancelReservationsByUserId(reservation).subscribe({
      next: (date) => {
        alert('Uspesno ste otkazali reservaciju');
        this.ngOnInit(); // reload strane
      },
      error: (err) => {
        alert(err.error.message);
      },
    });
  }
}
